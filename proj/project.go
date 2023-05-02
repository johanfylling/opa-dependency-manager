package proj

import (
	"crypto/sha256"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"styra.com/styrainc/odm/utils"
)

type Project struct {
	Name         string       `yaml:"name,omitempty"`
	Version      string       `yaml:"version,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
}

type DependencyInfo struct {
	Namespace string `yaml:"namespace,omitempty"`
	Version   string `yaml:"version,omitempty"`
}

type Dependency struct {
	DependencyInfo `yaml:",inline"`
	Location       string `yaml:"-"`
}

type Dependencies map[string]Dependency

func NewProject() *Project {
	return &Project{
		Dependencies: make(map[string]Dependency),
	}
}

func (ds *Dependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	infos := make(map[string]DependencyInfo)
	if err := unmarshal(&infos); err != nil {
		return err
	}

	*ds = make(map[string]Dependency)
	for k, v := range infos {
		(*ds)[k] = Dependency{
			DependencyInfo: v,
			Location:       k,
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

func (d Dependency) Update(rootDir string) error {
	var id string
	if d.Namespace != "" {
		id = d.Namespace
	} else {
		h := sha256.New()
		h.Write([]byte(d.Location))
		id = fmt.Sprintf("%x", h.Sum(nil))
	}

	sourceDir := fmt.Sprintf("%s/.", d.Location)
	targetDir := fmt.Sprintf("%s/%s", rootDir, id)

	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}

	if err := os.Mkdir(targetDir, 0755); err != nil {
		return err
	}

	if err := exec.Command("cp", "-a", sourceDir, targetDir).Run(); err != nil {
		return err
	}

	if d.Namespace != "" {
		mapping := fmt.Sprintf("data:data.%s", d.Namespace)
		if err := exec.Command("opa", "refactor", "move", "-w", "-p", mapping, targetDir).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) SetDependency(location string, info DependencyInfo) {
	if p.Dependencies == nil {
		p.Dependencies = make(map[string]Dependency)
	}
	p.Dependencies[location] = Dependency{
		DependencyInfo: info,
		Location:       location,
	}
}

func ReadProjectFromFile(path string) (*Project, error) {
	path = normalizeProjectPath(path)

	if !utils.FileExists(path) {
		return NewProject(), nil
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

	return &project, nil
}

func (p *Project) WriteToFile(path string, override bool) error {
	path = normalizeProjectPath(path)

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
