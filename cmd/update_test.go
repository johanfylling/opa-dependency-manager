package cmd

import (
	"github.com/johanfylling/odm/proj"
	"github.com/johanfylling/odm/utils"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUpdateProjects(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(file)

	tests := []struct {
		name          string
		projectDir    string
		expectedFiles map[string]*string
	}{
		{
			name:       "Empty project",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "empty"),
			expectedFiles: map[string]*string{
				filepath.Join(rootDir, "testdata", "projects", "empty", ".opa", "dependencies"): nil,
			},
		},
		{
			name:       "Project with source, no dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "no-dependencies"),
			expectedFiles: map[string]*string{
				filepath.Join(rootDir, "testdata", "projects", "no-dependencies", ".opa", "dependencies"): nil,
			},
		},
		{
			name:       "Project with multiple source dirs, no dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "source-list"),
			expectedFiles: map[string]*string{
				filepath.Join(rootDir, "testdata", "projects", "source-list", ".opa", "dependencies"): nil,
			},
		},
		{
			name:       "Project with local dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "local-dependencies"),
			expectedFiles: map[string]*string{
				filepath.Join(rootDir, "testdata", "projects", "local-dependencies", ".opa", "dependencies", proj.DepId("empty", "file:/../empty"), "opa.project"):             mustReadFile(rootDir, "testdata", "projects", "empty", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "local-dependencies", ".opa", "dependencies", proj.DepId("no_deps", "file:/../no-dependencies"), "opa.project"): mustReadFile(rootDir, "testdata", "projects", "no-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "local-dependencies", ".opa", "dependencies", proj.DepId("no_deps", "file:/../no-dependencies"), "src", "policy.rego"): ptr(`package no_deps.test

allow {
	1 + 1 == 2
}
`),
				filepath.Join(rootDir, "testdata", "projects", "local-dependencies", ".opa", "dependencies", proj.DepId("no_deps", "file:/../no-dependencies"), "tst", "tests.rego"): ptr(`package no_deps.test

test_allow {
	allow
}
`),
			},
		},
		{
			name:       "Project with local transitive dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies"),
			expectedFiles: map[string]*string{
				// foo
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("foo", "file:/../local-dependencies"), "opa.project"):      mustReadFile(rootDir, "testdata", "projects", "local-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("foo.empty", "file:/../empty"), "opa.project"):             mustReadFile(rootDir, "testdata", "projects", "empty", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("foo.no_deps", "file:/../no-dependencies"), "opa.project"): mustReadFile(rootDir, "testdata", "projects", "no-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("foo.no_deps", "file:/../no-dependencies"), "src", "policy.rego"): ptr(`package foo.no_deps.test

allow {
	1 + 1 == 2
}
`),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("foo.no_deps", "file:/../no-dependencies"), "tst", "tests.rego"): ptr(`package foo.no_deps.test

test_allow {
	allow
}
`),
				// bar
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("bar", "file:/../local-dependencies"), "opa.project"):      mustReadFile(rootDir, "testdata", "projects", "local-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("bar.empty", "file:/../empty"), "opa.project"):             mustReadFile(rootDir, "testdata", "projects", "empty", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("bar.no_deps", "file:/../no-dependencies"), "opa.project"): mustReadFile(rootDir, "testdata", "projects", "no-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("bar.no_deps", "file:/../no-dependencies"), "src", "policy.rego"): ptr(`package bar.no_deps.test

allow {
	1 + 1 == 2
}
`),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("bar.no_deps", "file:/../no-dependencies"), "tst", "tests.rego"): ptr(`package bar.no_deps.test

test_allow {
	allow
}
`),
				// baz
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("", "file:/../local-dependencies"), "opa.project"):     mustReadFile(rootDir, "testdata", "projects", "local-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("empty", "file:/../empty"), "opa.project"):             mustReadFile(rootDir, "testdata", "projects", "empty", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("no_deps", "file:/../no-dependencies"), "opa.project"): mustReadFile(rootDir, "testdata", "projects", "no-dependencies", "opa.project"),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("no_deps", "file:/../no-dependencies"), "src", "policy.rego"): ptr(`package no_deps.test

allow {
	1 + 1 == 2
}
`),
				filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies", ".opa", "dependencies", proj.DepId("no_deps", "file:/../no-dependencies"), "tst", "tests.rego"): ptr(`package no_deps.test

test_allow {
	allow
}
`),
			},
		},
	}

	for _, tc := range tests {
		//goland:noinspection GoDeferInLoop
		defer cleanup(tc.projectDir)

		t.Run(tc.name, func(t *testing.T) {
			if err := doUpdate(tc.projectDir); err != nil {
				t.Fatal(err)
			}

			for filePath, expected := range tc.expectedFiles {
				if !utils.FileExists(filePath) {
					t.Fatalf("expected file '%s' to exist", filePath)
				}

				if expected != nil {
					b, err := os.ReadFile(filePath)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Compare(string(b), *expected) != 0 {
						t.Fatalf("expected file '%s' to contain:\n\n%s\n\ngot:\n\n%s", filePath, *expected, string(b))
					}
				} else {
					children, err := os.ReadDir(filePath)
					if err != nil {
						t.Fatal(err)
					}
					if len(children) != 0 {
						t.Fatalf("expected dir '%s' to be empty, got %d children", filePath, len(children))
					}
				}
			}
		})
	}
}

func ptr(s string) *string {
	return &s
}

func readFiles(t *testing.T, paths ...string) map[string]*string {
	t.Helper()
	files := make(map[string]*string)
	for _, path := range paths {
		p, v := readFile(t, path)
		files[p] = v
	}
	return files
}

func readFile(t *testing.T, path string) (string, *string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	return path, &s
}

func mustReadFile(path ...string) *string {
	p := filepath.Join(path...)
	b, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return &s
}

func cleanup(projectDir string) {
	dotOpaDir := filepath.Join(projectDir, ".opa")
	_ = os.RemoveAll(dotOpaDir)
}
