package proj

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/utils"
	"path/filepath"
	"strings"
)

type Location string

func (l *Location) MarshalYAML() (interface{}, error) {
	return string(*l), nil
}

func (l *Location) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var location string
	if err := unmarshal(&location); err != nil {
		return err
	}
	*l = Location(location)
	return nil
}

func (l *Location) String() string {
	return string(*l)
}

func (l *Location) Type() string {
	if l.IsLocal() {
		return "local"
	} else if l.IsGit() {
		return "git"
	}
	return "unknown"
}

func (l *Location) IsLocal() bool {
	return strings.HasPrefix(string(*l), "file:")
}

func (l *Location) IsGit() bool {
	return strings.HasPrefix(string(*l), "git:") || strings.HasPrefix(string(*l), "git+")
}

func (l *Location) IsSupported() bool {
	return l.IsLocal() || l.IsGit()
}

func (l *Location) Clone(rootDir, targetDir string) error {
	if l.IsLocal() {
		return l.cloneLocal(rootDir, targetDir)
	} else if l.IsGit() {
		return l.cloneGit(targetDir)
	}
	return fmt.Errorf("unsupported location type: %s", l)
}

func (l *Location) cloneLocal(rootDir, targetDir string) error {
	sourceDir, err := utils.NormalizeFilePath(string(*l))
	if err != nil {
		return err
	}

	if !filepath.IsAbs(sourceDir) {
		sourceDir = filepath.Join(rootDir, sourceDir)
	}

	if !utils.FileExists(sourceDir) {
		return fmt.Errorf("source dir %s does not exist", sourceDir)
	}

	if !utils.IsDir(sourceDir) && utils.GetFileName(sourceDir) == "opa.project" {
		sourceDir = utils.GetParentDir(sourceDir)
	}

	// Ignore empty files, as an empty module will break the 'opa refactor' command
	if err := utils.CopyAll(sourceDir, targetDir, []string{".opa"}, true); err != nil {
		return err
	}

	return nil
}

func (l *Location) cloneGit(targetDir string) error {
	url, tag, err := parseGitUrl(string(*l))
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
