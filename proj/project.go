package proj

import (
	"crypto/sha256"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/utils"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	dotOpaDir = ".opa"
	depDir    = "dependencies"
)

type Project struct {
	Name         string       `yaml:"name,omitempty"`
	Version      string       `yaml:"version,omitempty"`
	SourceDirs   []string     `yaml:"source,omitempty"`
	TestDirs     []string     `yaml:"tests,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
	Build        Build        `yaml:"build,omitempty"`
	filePath     string
}

type ProjectSerialization struct {
	Name         string       `yaml:"name,omitempty"`
	Version      string       `yaml:"version,omitempty"`
	Source       interface{}  `yaml:"source,omitempty"`
	Test         interface{}  `yaml:"tests,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
	Build        Build        `yaml:"build,omitempty"`
}

type Build struct {
	Output      string   `yaml:"output,omitempty"`
	Target      string   `yaml:"target,omitempty"`
	Entrypoints []string `yaml:"entrypoints,omitempty"`
}

type DependencyInfo struct {
	Location  string `yaml:"location"`
	Namespace string `yaml:"namespace,omitempty"`
}

type Dependency struct {
	DependencyInfo   `yaml:",inline"`
	Name             string      `yaml:"-"`
	Project          *Project    `yaml:"-"`
	ParentDependency *Dependency `yaml:"-"`
	dirPath          string      `yaml:"-"`
}

type Dependencies map[string]Dependency

func NewProject(path string) *Project {
	return &Project{
		Dependencies: make(map[string]Dependency),
		filePath:     path,
	}
}

func (ds *Dependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	raw := make(map[string]interface{})
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*ds = make(map[string]Dependency)
	for k, v := range raw {
		var info DependencyInfo
		switch v.(type) {
		case string:
			info = DependencyInfo{
				Location:  v.(string),
				Namespace: k,
			}
		case map[string]interface{}:
			var namespace = ""
			if ns := v.(map[string]interface{})["namespace"]; ns != nil {
				switch ns := ns.(type) {
				case bool:
					if ns {
						namespace = k
					}
				case string:
					namespace = ns
				default:
					return fmt.Errorf("invalid namespace type: %T", ns)
				}
			} else {
				// If no namespace is specified, default to the dependency name
				namespace = k
			}
			info = DependencyInfo{
				Location:  v.(map[string]interface{})["location"].(string),
				Namespace: namespace,
			}
		}
		(*ds)[k] = Dependency{
			DependencyInfo: info,
			Name:           k,
		}
	}

	return nil
}

func (ds *Dependencies) MarshalYAML() (interface{}, error) {
	depMap := make(map[string]Dependency)
	for _, dep := range *ds {
		depMap[dep.Location] = dep
	}

	return depMap, nil
}

func (d Dependency) MarshalYAML() (interface{}, error) {
	printer.Debug("Marshalling dependency %s", d.Name)

	if d.Namespace == d.Name {
		return d.Location, nil
	}

	if d.Namespace == "" {
		return map[string]interface{}{
			"namespace": false,
			"location":  d.Location,
		}, nil

	}

	return map[string]interface{}{
		"namespace": d.Namespace,
		"location":  d.Location,
	}, nil
}

func (d Dependency) id() string {
	return DepId(d.fullNamespace(), d.Location)
}

func DepId(namespace, location string) string {
	cleartext := fmt.Sprintf("%s:%s", namespace, location)
	h := sha256.New()
	h.Write([]byte(cleartext))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (d Dependency) dir(rootDir string) string {
	return filepath.Join(rootDir, d.id())
}

func (d Dependency) Update(rootDir, depsRootDir string) error {
	targetDir := d.dir(depsRootDir)

	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", targetDir, err)
	}

	if strings.HasPrefix(d.Location, "git+") {
		printer.Debug("Updating git dependency %s", d.Namespace)
		if err := d.updateGit(targetDir); err != nil {
			return err
		}
	} else if strings.HasPrefix(d.Location, "file:") {
		printer.Debug("Updating git dependency %s", d.Namespace)
		printer.Debug("Updating transitive dependencies for %s", d.Namespace)
		if err := d.updateLocal(rootDir, targetDir); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported dependency location: %s", d.Location)
	}

	depProjectFile := fmt.Sprintf("%s/opa.project", targetDir)
	if utils.FileExists(depProjectFile) {
		var err error
		d.Project, err = ReadProjectFromFile(depProjectFile, false)
		if err != nil {
			return err
		}
	}
	d.dirPath = targetDir

	if err := d.updateTransitive(rootDir, depsRootDir); err != nil {
		return fmt.Errorf("failed to update transitive dependencies for %s: %w", d.Namespace, err)
	}

	if namespace := d.fullNamespace(); namespace != "" {
		var dirs []string
		if srcDirs := d.SourceDirs(); len(srcDirs) > 0 {
			dirs = append(dirs, srcDirs...)
		} else {
			dirs = append(dirs, targetDir)
		}
		dirs = append(dirs, d.TestDirs()...)
		dirs = utils.FilterExistingFiles(dirs)

		if len(dirs) > 0 {
			opa := utils.NewOpa(dirs...)
			if err := opa.Refactor("data", fmt.Sprintf("data.%s", namespace)); err != nil {
				return fmt.Errorf("failed to refactor namespace %s: %w", d.Namespace, err)
			}
		} else {
			printer.Debug("Dependency %s has no source, skipping namespace refactoring", d.Name)
		}
	}

	return nil
}

func (d Dependency) Load(rootDir, targetDir string) (*Dependency, error) {
	targetDir = d.dir(targetDir)
	d.dirPath = targetDir
	depProjectFile := fmt.Sprintf("%s/opa.project", targetDir)
	if utils.FileExists(depProjectFile) {
		var err error
		d.Project, err = ReadProjectFromFile(depProjectFile, false)
		if err != nil {
			return nil, err
		}
	}
	if err := d.loadTransitive(rootDir, targetDir); err != nil {
		return nil, fmt.Errorf("failed to update transitive dependencies for %s: %w", d.Namespace, err)
	}
	return &d, nil
}

func (d Dependency) updateLocal(rootDir, targetDir string) error {
	sourceLocation, err := utils.NormalizeFilePath(d.Location)
	if err != nil {
		return err
	}

	if !filepath.IsAbs(sourceLocation) {
		sourceLocation = filepath.Join(rootDir, sourceLocation)
	}

	if !utils.FileExists(sourceLocation) {
		return fmt.Errorf("dependency %s does not exist", sourceLocation)
	}

	if !utils.IsDir(sourceLocation) && utils.GetFileName(sourceLocation) == "opa.project" {
		sourceLocation = utils.GetParentDir(sourceLocation)
	}

	// Ignore empty files, as an empty module will break the 'opa refactor' command
	if err := utils.CopyAll(sourceLocation, targetDir, []string{".opa"}, true); err != nil {
		return err
	}

	return nil
}

func (d Dependency) updateGit(targetDir string) error {
	url, tag, err := parseGitUrl(d.Location)
	if err != nil {
		return err
	}

	repo, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      url,
		Progress: printer.DebugPrinter(),
	})
	if err != nil {
		return fmt.Errorf("failed to clone git repository %s: %w", url, err)
	}

	if tag != "" {
		w, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree for git repository %s: %w", url, err)
		}

		if err := w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(tag),
		}); err != nil {
			return fmt.Errorf("failed to checkout tag '%s' for git repository %s: %w", tag, url, err)
		}
	} else {
		printer.Debug("No tag specified, using HEAD")
	}

	return nil
}

func parseGitUrl(fullUrl string) (url string, tag string, err error) {
	trimmedUrl := strings.TrimPrefix(fullUrl, "git+")
	parts := strings.Split(trimmedUrl, "#")
	if len(parts) > 2 {
		return "", "", fmt.Errorf("invalid git url %s; only one tag separator '#' allowed", fullUrl)
	}

	url = parts[0]
	if len(parts) == 2 {
		tag = parts[1]
	}
	return
}

func (d Dependency) loadTransitive(rootDir, targetDir string) error {
	printer.Debug("Loading transitive dependencies for %s (%s)", d.Namespace, d.id())

	if d.Project != nil {
		for i, dep := range d.Project.Dependencies {
			if dep, err := dep.Load(rootDir, targetDir); err != nil {
				return err
			} else {
				dep.ParentDependency = &d
				d.Project.Dependencies[i] = *dep
			}
		}
	}

	return nil
}

func (d Dependency) updateTransitive(rootDir, targetDir string) error {
	printer.Debug("Updating transitive dependencies for %s (%s)", d.Namespace, d.id())

	if d.Project != nil {
		for name, dep := range d.Project.Dependencies {
			dep.ParentDependency = &d
			if err := dep.Update(rootDir, targetDir); err != nil {
				return err
			}
			d.Project.Dependencies[name] = dep
		}
	}

	return nil
}

func (d Dependency) fullNamespace() string {
	if d.ParentDependency != nil && d.ParentDependency.Namespace != "" {
		if parentNamespace := d.ParentDependency.fullNamespace(); parentNamespace != "" {
			if d.Namespace == "" {
				return parentNamespace
			}
			return fmt.Sprintf("%s.%s", parentNamespace, d.Namespace)
		}
	}
	return d.Namespace
}

func (d Dependency) SourceDirs() []string {
	if d.Project != nil && len(d.Project.SourceDirs) > 0 {
		dirs := make([]string, 0, len(d.Project.SourceDirs))
		for _, dir := range d.Project.SourceDirs {
			dirs = append(dirs, filepath.Join(d.dirPath, dir))
		}
		return dirs
	}
	return []string{d.dirPath}
}

func (d Dependency) TestDirs() []string {
	if d.Project != nil && len(d.Project.TestDirs) > 0 {
		dirs := make([]string, 0, len(d.Project.TestDirs))
		for _, dir := range d.Project.TestDirs {
			dirs = append(dirs, filepath.Join(d.dirPath, dir))
		}
		return dirs
	}
	return []string{}
}

func (p *Project) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw ProjectSerialization
	if err := unmarshal(&raw); err != nil {
		return err
	}

	p.Name = raw.Name
	p.Version = raw.Version
	p.Dependencies = raw.Dependencies
	p.Build = raw.Build

	var err error
	p.SourceDirs, err = unmarshalDirs(raw.Source)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	p.TestDirs, err = unmarshalDirs(raw.Test)
	if err != nil {
		return fmt.Errorf("invalid tests: %w", err)
	}

	return nil
}

func (p Project) MarshalYAML() (interface{}, error) {
	var raw ProjectSerialization
	raw.Name = p.Name
	raw.Version = p.Version
	raw.Dependencies = p.Dependencies
	raw.Build = p.Build
	if len(p.SourceDirs) == 1 {
		raw.Source = p.SourceDirs[0]
	} else if len(p.SourceDirs) > 1 {
		raw.Source = p.SourceDirs
	}
	if len(p.TestDirs) == 1 {
		raw.Test = p.TestDirs[0]
	} else if len(p.TestDirs) > 1 {
		raw.Test = p.TestDirs
	}
	return raw, nil
}

func unmarshalDirs(raw interface{}) ([]string, error) {
	switch t := raw.(type) {
	case nil:
		return nil, nil
	case string:
		return []string{t}, nil
	case []string:
		return t, nil
	case []interface{}:
		dirs := make([]string, len(t))
		for i, v := range t {
			var ok bool
			if dirs[i], ok = v.(string); !ok {
				return nil, fmt.Errorf("invalid dir type %T", v)
			}
		}
		return dirs, nil
	default:
		return nil, fmt.Errorf("invalid dir type %T", t)
	}
}

func (p *Project) SetDependency(name string, info DependencyInfo) {
	if p.Dependencies == nil {
		p.Dependencies = make(map[string]Dependency)
	}
	p.Dependencies[name] = Dependency{
		DependencyInfo: info,
		Name:           name,
	}
}

func ReadProjectFromFile(path string, allowMissing bool) (*Project, error) {
	path = normalizeProjectPath(path)

	if !utils.FileExists(path) {
		if allowMissing {
			return NewProject(path), nil
		} else {
			return nil, fmt.Errorf("project file %s does not exist", path)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project file %s: %w", path, err)
	}

	var project Project
	err = yaml.Unmarshal(data, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project file %s: %w", path, err)
	}

	project.filePath = path

	return &project, nil
}

func ReadAndLoadProject(path string, allowMissing bool) (*Project, error) {
	project, err := ReadProjectFromFile(path, allowMissing)
	if err != nil {
		return nil, err
	}

	if err := project.Load(); err != nil {
		return nil, err
	}

	return project, nil
}

func (p *Project) Update() error {
	rootDir := filepath.Dir(p.filePath)
	return p.update(rootDir)
}

func (p *Project) update(rootDir string) error {
	depRootDir := dependenciesDir(rootDir)

	for name, dep := range p.Dependencies {
		if err := dep.Update(rootDir, depRootDir); err != nil {
			return fmt.Errorf("failed to update dependency %s: %w", name, err)
		}
		p.Dependencies[name] = dep
	}

	return nil
}

func (p *Project) Load() error {
	rootDir := filepath.Dir(p.filePath)
	return p.load(rootDir)
}

func (p *Project) load(rootDir string) error {
	depRootDir := dependenciesDir(rootDir)

	for name, dep := range p.Dependencies {
		// Load, don't update dependencies, this is done separately
		if loadedDep, err := dep.Load(rootDir, depRootDir); err != nil {
			return fmt.Errorf("failed to load dependency %s: %w", name, err)
		} else {
			dep = *loadedDep
		}
		if dep.Project != nil {
			if err := dep.Project.load(rootDir); err != nil {
				return fmt.Errorf("failed loading dependency project: %w", err)
			}
		}
		p.Dependencies[name] = dep
	}

	return nil
}

func (p *Project) DataLocations() ([]string, error) {
	var dataLocations []string
	projDir := filepath.Dir(p.filePath)
	if len(p.SourceDirs) > 0 {
		for _, dir := range p.SourceDirs {
			if dir, err := utils.NormalizeFilePath(dir); err != nil {
				return nil, err
			} else {
				dataLocations = append(dataLocations, filepath.Join(projDir, dir))
			}
		}
	} else {
		dataLocations = append(dataLocations, projDir)
	}

	err := WalkDependencies(p, func(dep Dependency) error {
		dataLocations = append(dataLocations, dep.SourceDirs()...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	dataLocations = utils.FilterExistingFiles(dataLocations)

	return dataLocations, nil
}

func (p *Project) TestLocations(includeDependencies bool) ([]string, error) {
	var testLocations []string
	projDir := filepath.Dir(p.filePath)
	if len(p.TestDirs) > 0 {
		for _, dir := range p.TestDirs {
			if dir, err := utils.NormalizeFilePath(dir); err != nil {
				return nil, err
			} else {
				testLocations = append(testLocations, filepath.Join(projDir, dir))
			}
		}
	}

	if includeDependencies {
		err := WalkDependencies(p, func(dep Dependency) error {
			testLocations = append(testLocations, dep.TestDirs()...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return testLocations, nil
}

func (p *Project) WriteToFile(path string, override bool) error {
	path = normalizeProjectPath(path)
	printer.Debug("Writing project file to %s", path)

	if !override && utils.FileExists(path) {
		return fmt.Errorf("project file %s already exists", path)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal project file %s: %w", path, err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write project file %s: %w", path, err)
	}

	return nil
}

func (p *Project) PrintTree(w io.Writer) error {
	if err := p.printTree(w, "root", 0); err != nil {
		return err
	}
	return nil
}

func (p *Project) printTree(w io.Writer, name string, indent int) error {
	indentStr := strings.Repeat(" ", indent*2)
	if p == nil {
		_, err := fmt.Fprintf(w, "%s%s\n", indentStr, name)
		return err
	}

	if len(p.Name) > 0 {
		if _, err := fmt.Fprintf(w, "%s%s (%s)\n", indentStr, name, p.Name); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "%s%s\n", indentStr, name); err != nil {
			return err
		}
	}
	for _, dep := range p.Dependencies {
		if err := dep.Project.printTree(w, dep.Name, indent+1); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) Dir() string {
	return filepath.Dir(p.filePath)
}

func normalizeProjectPath(path string) string {
	l := len(path)
	if l >= 11 && path[l-11:] == "opa.project" {
		return path
	} else if l >= 1 && path[l-1] == '/' {
		return path + "opa.project"
	} else {
		return path + "/opa.project"
	}
}

func dependenciesDir(root string) string {
	return filepath.Join(root, dotOpaDir, depDir)
}

func WalkDependencies(p *Project, f func(Dependency) error) error {
	if p == nil {
		return nil
	}

	for _, dep := range p.Dependencies {
		if err := f(dep); err != nil {
			return err
		}
		if dep.Project != nil {
			if err := WalkDependencies(dep.Project, f); err != nil {
				return err
			}
		}
	}

	return nil
}
