package proj

import (
	"crypto/sha256"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/utils"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Project struct {
	Name         string       `yaml:"name,omitempty"`
	Version      string       `yaml:"version,omitempty"`
	SourceDir    string       `yaml:"source,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
}

type DependencyInfo struct {
	Namespace string `yaml:"namespace,omitempty"`
	Version   string `yaml:"version,omitempty"`
}

type Dependency struct {
	DependencyInfo         `yaml:",inline"`
	Location               string       `yaml:"-"`
	TransitiveDependencies Dependencies `yaml:"-"`
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

func (d *Dependency) Update(rootDir string) error {
	var id string
	if d.Namespace != "" {
		id = d.Namespace
	} else {
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
	} else { //if strings.HasPrefix(d.Location, "file:") {
		printer.Debug("Updating git dependency %s", d.Namespace)
		printer.Debug("Updating transitive dependencies for %s", d.Namespace)
		if err := d.updateLocal(targetDir); err != nil {
			return err
		}
	} //} else {
	//	return fmt.Errorf("unsupported dependency location: %s", d.Location)
	//}

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

func (d *Dependency) updateLocal(targetDir string) error {
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

func (d *Dependency) updateGit(targetDir string) error {
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

func (d *Dependency) updateTransitive(targetDir string) error {
	printer.Debug("Updating transitive dependencies for %s", d.Namespace)

	transitiveProjectFile := fmt.Sprintf("%s/opa.project", targetDir)
	if utils.FileExists(transitiveProjectFile) {
		transitiveProject, err := ReadProjectFromFile(transitiveProjectFile, false)
		if err != nil {
			return err
		}

		transitiveRootDir := fmt.Sprintf("%s/_transitive", targetDir)
		for _, dep := range transitiveProject.Dependencies {
			if err := dep.Update(transitiveRootDir); err != nil {
				return err
			}
		}

		d.TransitiveDependencies = transitiveProject.Dependencies
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

func ReadProjectFromFile(path string, allowMissing bool) (*Project, error) {
	path = normalizeProjectPath(path)

	if !utils.FileExists(path) {
		if allowMissing {
			return NewProject(), nil
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

	return &project, nil
}

func (p *Project) DataLocations() ([]string, error) {
	dataLocations := []string{"./.opa/dependencies"}
	if p.SourceDir != "" {
		if src, err := utils.NormalizeFilePath(p.SourceDir); err != nil {
			return nil, err
		} else {
			dataLocations = append(dataLocations, src)
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
