package proj

import (
	"crypto/sha256"
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Repository struct {
	RepositorySerialization
	Location Location
}

type RepositorySerialization struct {
	Libraries Libraries `yaml:"libraries"`
}

type Library Location

type Libraries map[string]Library

func (ls Libraries) Find(name string) *Location {
	l := ls[name]
	return (*Location)(&l)
}

func (r *Repository) update(rootDir, repoRootDir string) error {
	targetDir := r.dir(repoRootDir)

	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	printer.Debug("Updating repository: %s", r.Location)
	if !r.Location.IsSupported() {
		return fmt.Errorf("unsupported location type: %s", r.Location.String())
	}
	if err := r.Location.Clone(rootDir, targetDir); err != nil {
		return err
	}

	repoFile := fmt.Sprintf("%s/repository.yaml", targetDir)
	if !utils.FileExists(repoFile) {
		return fmt.Errorf("no repository file")
	}

	if repo, err := ReadRepositoryFromFile(repoFile); err != nil {
		return err
	} else {
		r.Libraries = repo.Libraries
	}

	return nil
}

func (r *Repository) dir(rootDir string) string {
	return filepath.Join(rootDir, r.id())
}

func (r *Repository) id() string {
	cleartext := r.Location.String()
	h := sha256.New()
	h.Write([]byte(cleartext))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ReadRepositoryFromFile(path string) (*RepositorySerialization, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project file %s: %w", path, err)
	}

	var repo RepositorySerialization
	err = yaml.Unmarshal(data, &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal repository file %s: %w", path, err)
	}
	return &repo, nil
}
