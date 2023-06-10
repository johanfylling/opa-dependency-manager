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
	SourceDir    string       `yaml:"source,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
	Build        Build        `yaml:"build,omitempty"`
	filePath     string
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
	DependencyInfo `yaml:",inline"`
	Name           string   `yaml:"-"`
	Project        *Project `yaml:"-"`
	dirPath        string   `yaml:"-"`
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

func (d Dependency) Update(rootDir string) error {
	var id string
	if d.Name != "" {
		id = d.Name
	} else {
		// Should never happen, as dependencies are keyed by their name
		h := sha256.New()
		h.Write([]byte(d.Location))
		id = fmt.Sprintf("%x", h.Sum(nil))
	}

	targetDir := fmt.Sprintf("%s/%s", rootDir, id)

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
		if err := d.updateLocal(targetDir); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported dependency location: %s", d.Location)
	}

	if err := d.updateTransitive(targetDir); err != nil {
		return fmt.Errorf("failed to update transitive dependencies for %s: %w", d.Namespace, err)
	}

	if d.Namespace != "" {
		mapping := fmt.Sprintf("data:data.%s", d.Namespace)
		if _, err := utils.RunCommand("opa", "refactor", "move", "-w", "-p", mapping, targetDir); err != nil {
			return fmt.Errorf("failed to refactor namespace %s: %w", d.Namespace, err)
		}
	}

	return nil
}

func (d Dependency) updateLocal(targetDir string) error {
	sourceLocation, err := utils.NormalizeFilePath(d.Location)
	if err != nil {
		return err
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

func (d Dependency) updateTransitive(targetDir string) error {
	printer.Debug("Updating transitive dependencies for %s", d.Namespace)

	transitiveProjectFile := fmt.Sprintf("%s/opa.project", targetDir)
	if utils.FileExists(transitiveProjectFile) {
		transitiveProject, err := ReadProjectFromFile(transitiveProjectFile, false)
		if err != nil {
			return err
		}

		transitiveRootDir := dependenciesDir(targetDir)
		for _, dep := range transitiveProject.Dependencies {
			if err := dep.Update(transitiveRootDir); err != nil {
				return err
			}
		}

		d.Project = transitiveProject
	}

	return nil
}

func (d Dependency) SourceDir() string {
	if d.Project != nil && d.Project.SourceDir != "" {
		return filepath.Join(d.dirPath, d.Project.SourceDir)
	}
	return d.dirPath
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

func (p *Project) Load() error {
	rootDir := filepath.Dir(p.filePath)
	depRootDir := dependenciesDir(rootDir)

	for name, dep := range p.Dependencies {
		dep.dirPath = filepath.Join(depRootDir, name)
		depProjFile := normalizeProjectPath(dep.dirPath)
		if utils.FileExists(depProjFile) {
			var err error
			if dep.Project, err = ReadProjectFromFile(depProjFile, true); err != nil {
				return fmt.Errorf("failed reading dependency project from file: %w", err)
			}
			if err := dep.Project.Load(); err != nil {
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
	if p.SourceDir != "" {
		if src, err := utils.NormalizeFilePath(p.SourceDir); err != nil {
			return nil, err
		} else {
			dataLocations = append(dataLocations, filepath.Join(projDir, src))
		}
	} else {
		dataLocations = append(dataLocations, projDir)
	}

	for _, dep := range p.Dependencies {
		if depProject := dep.Project; depProject != nil {
			if childLocations, err := depProject.DataLocations(); err != nil {
				return nil, err
			} else {
				dataLocations = append(dataLocations, childLocations...)
			}
		} else {
			dataLocations = append(dataLocations, dep.SourceDir())
		}
	}

	return dataLocations, nil
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
