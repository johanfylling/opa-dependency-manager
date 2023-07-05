package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/proj"
	"github.com/johanfylling/odm/utils"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildProjects(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(file)

	tests := []struct {
		name           string
		projectDir     string
		bundleLocation string
		bundleContent  []string
		cleanup        string
	}{
		{
			name:           "Empty project",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "empty"),
			bundleLocation: filepath.Join(rootDir, "testdata", "projects", "empty", "build", "bundle.tar.gz"),
			cleanup:        "build",
			bundleContent: []string{
				"/data.json",
			},
		},
		{
			name:           "Project with source, no dependencies",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "no-dependencies"),
			bundleLocation: filepath.Join(rootDir, "testdata", "projects", "no-dependencies", "build", "bundle.tar.gz"),
			cleanup:        "build",
			bundleContent: []string{
				"/data.json",
				"/src/policy.rego",
			},
		},
		{
			name:           "Project with multiple source dirs, no dependencies",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "source-list"),
			bundleLocation: filepath.Join(rootDir, "testdata", "projects", "source-list", "build", "bundle.tar.gz"),
			cleanup:        "build",
			bundleContent: []string{
				"/data.json",
				"/src/policy.rego",
			},
		},
		{
			name:           "Project with local dependencies",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "local-dependencies"),
			bundleLocation: filepath.Join(rootDir, "testdata", "projects", "local-dependencies", "build", "bundle.tar.gz"),
			cleanup:        "build",
			bundleContent: []string{
				"/data.json",
				filepath.Join(proj.DepId("no_deps", "file:/../no-dependencies"), "src", "policy.rego"),
			},
		},
		{
			name:           "Project with transitive dependencies",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies"),
			bundleLocation: filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", "build", "bundle.tar.gz"),
			cleanup:        "build",
			bundleContent: []string{
				"/data.json",
				filepath.Join(proj.DepId("foo.no_deps", "file:/../no-dependencies"), "src", "policy.rego"),
				filepath.Join(proj.DepId("bar.no_deps", "file:/../no-dependencies"), "src", "policy.rego"),
				filepath.Join(proj.DepId("no_deps", "file:/../no-dependencies"), "src", "policy.rego"),
			},
		},
	}

	for _, tc := range tests {
		//goland:noinspection GoDeferInLoop
		defer cleanup(tc.projectDir, tc.cleanup)

		t.Run(tc.name, func(t *testing.T) {
			output := bytes.Buffer{}
			printer.PrintWriter = &output
			args := []string{}
			if err := doUpdate(tc.projectDir); err != nil {
				t.Fatal(err)
			}
			if err := doBuild(tc.projectDir, args); err != nil {
				t.Fatal(err)
			}
			if !utils.FileExists(tc.bundleLocation) {
				t.Fatalf("expected bundle file to exist at %s", tc.bundleLocation)
			}

			var bundleFiles []string
			if r, err := os.Open(tc.bundleLocation); err != nil {
				t.Fatal(err)
			} else {
				uncompressedStream, err := gzip.NewReader(r)
				if err != nil {
					t.Fatal(err)
				}
				reader := tar.NewReader(uncompressedStream)
				for {
					header, err := reader.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					bundleFiles = append(bundleFiles, header.Name)
				}
			}
			if len(bundleFiles) != len(tc.bundleContent) {
				t.Fatalf("expected files in bundle:\n\n%vgot:\n\n%v", tc.bundleContent, bundleFiles)
			}
			for _, expectedFile := range tc.bundleContent {
				found := false
				for _, actualFile := range bundleFiles {
					// tared file header name is full original source path, for some reason, so we can't do full match
					if strings.HasSuffix(actualFile, expectedFile) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected file %s to be in bundle, but it wasn't", expectedFile)
				}
			}
		})
	}
}
