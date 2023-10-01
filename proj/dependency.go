package proj

import (
	"crypto/sha256"
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/utils"
	"os"
	"path/filepath"
	"strings"
)

type Dependency struct {
	DependencyInfo   `yaml:",inline"`
	Name             string      `yaml:"-"`
	Project          *Project    `yaml:"-"`
	ParentDependency *Dependency `yaml:"-"`
	dirPath          string      `yaml:"-"`
}

type DependencyInfo struct {
	Location  Location `yaml:"location"`
	Namespace string   `yaml:"namespace,omitempty"`
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
	return DepId(d.fullNamespace(), d.Location.String())
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

func (d Dependency) Update(parentProject *Project, rootDir, depsRootDir string) error {
	targetDir := d.dir(depsRootDir)

	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	printer.Debug("Updating %s dependency: %s", d.Location.Type(), d.Namespace)
	if !d.Location.IsSupported() {
		var lib *Location
		for _, repo := range parentProject.Repositories {
			if lib = repo.Libraries.Find(d.Location.String()); lib != nil {
				break
			}
		}
		if lib == nil {
			return fmt.Errorf("unsupported location type: %s", d.Location.String())
		}
		d.Location = *lib
	}
	if err := d.Location.Clone(rootDir, targetDir); err != nil {
		return err
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

		//namespace = strings.ReplaceAll(namespace, "-", "_")

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

	if p := d.Project; p != nil {
		if len(p.Repositories) > 0 {
			repoRootDir := repositoriesDir(rootDir)
			for i, repo := range p.Repositories {
				if err := repo.update(rootDir, repoRootDir); err != nil {
					return fmt.Errorf("failed to update repository %d: %w", i+1, err)
				}
				p.Repositories[i] = repo
			}
		}

		for name, dep := range p.Dependencies {
			dep.ParentDependency = &d
			if err := dep.Update(p, rootDir, targetDir); err != nil {
				return err
			}
			p.Dependencies[name] = dep
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
