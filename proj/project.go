package proj

import (
	"fmt"
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
	repoDir   = "repositories"
)

type Project struct {
	Name         string
	Version      string
	SourceDirs   []string
	TestDirs     []string
	Dependencies Dependencies
	Build        Build
	Repositories []Repository
	filePath     string
}

type ProjectSerialization struct {
	Name         string       `yaml:"name,omitempty"`
	Version      string       `yaml:"version,omitempty"`
	Source       interface{}  `yaml:"source,omitempty"`
	Test         interface{}  `yaml:"tests,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
	Build        Build        `yaml:"build,omitempty"`
	Repositories []string     `yaml:"repositories,omitempty"`
}

type Build struct {
	Output      string   `yaml:"output,omitempty"`
	Target      string   `yaml:"target,omitempty"`
	Entrypoints []string `yaml:"entrypoints,omitempty"`
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
		switch v := v.(type) {
		case string:
			info = DependencyInfo{
				Location:  Location(v),
				Namespace: k,
			}
		case map[string]interface{}:
			var namespace = ""
			if ns := v["namespace"]; ns != nil {
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
				Location:  Location(v["location"].(string)),
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
		depMap[dep.Location.String()] = dep
	}

	return depMap, nil
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

	p.Repositories, err = unmarshalRepositories(raw.Repositories)
	if err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	return nil
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

func unmarshalRepositories(raw interface{}) ([]Repository, error) {
	switch t := raw.(type) {
	case nil:
		return nil, nil
	case string:
		return []Repository{{Location: Location(t)}}, nil
	case []string:
		repos := make([]Repository, len(t))
		for i, v := range t {
			repos[i] = Repository{Location: Location(v)}
		}
		return repos, nil
	default:
		return nil, fmt.Errorf("invalid repository type %T", t)
	}
}

func (p *Project) MarshalYAML() (interface{}, error) {
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

	raw.Repositories = make([]string, len(p.Repositories))
	for _, repo := range p.Repositories {
		if repoYaml, err := repo.Location.MarshalYAML(); err != nil {
			return nil, err
		} else {
			raw.Repositories = append(raw.Repositories, repoYaml.(string))
		}
	}

	return raw, nil
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
	if len(p.Repositories) > 0 {
		repoRootDir := repositoriesDir(rootDir)
		for i, repo := range p.Repositories {
			if err := repo.update(rootDir, repoRootDir); err != nil {
				return fmt.Errorf("failed to update repository %d: %w", i+1, err)
			}
			p.Repositories[i] = repo
		}
	}

	depRootDir := dependenciesDir(rootDir)

	for name, dep := range p.Dependencies {
		if err := dep.Update(p, rootDir, depRootDir); err != nil {
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

func repositoriesDir(root string) string {
	return filepath.Join(root, dotOpaDir, repoDir)
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
