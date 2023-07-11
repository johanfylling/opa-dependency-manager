// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package proj

import (
	"crypto/sha256"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	Location  string `yaml:"location"`
	Namespace string `yaml:"namespace,omitempty"`
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
		if err := d.updateFromGit(targetDir); err != nil {
			return err
		}
	} else if strings.HasPrefix(d.Location, "file:") {
		printer.Debug("Updating git dependency %s", d.Namespace)
		printer.Debug("Updating transitive dependencies for %s", d.Namespace)
		if err := d.updateFromLocal(rootDir, targetDir); err != nil {
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

func (d Dependency) updateFromLocal(rootDir, targetDir string) error {
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

func (d Dependency) updateFromGit(targetDir string) error {
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
