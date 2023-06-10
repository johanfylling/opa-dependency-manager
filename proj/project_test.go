package proj

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMarshalProject(t *testing.T) {
	tests := []struct {
		note     string
		project  *Project
		expected string
	}{
		{
			note: "no dependencies",
			project: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
			},
			expected: `name: test_project
version: 0.0.1
source: src
`,
		},
		{
			note: "file dependency with only name & location",
			project: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "foo",
						},
					},
				},
			},
			expected: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo: file://dev/null
`,
		},
		{
			note: "file dependency with no namespace",
			project: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "",
						},
					},
				},
			},
			expected: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo:
        location: file://dev/null
        namespace: false
`,
		},
		{
			note: "file dependency with named namespace",
			project: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "bar",
						},
					},
				},
			},
			expected: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo:
        location: file://dev/null
        namespace: bar
`,
		},
		{
			note: "git dependency with only name & location",
			project: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "git+https://example.com/my/repo",
							Namespace: "foo",
						},
					},
				},
			},
			expected: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo: git+https://example.com/my/repo
`,
		},
	}

	for _, test := range tests {
		t.Run(test.note, func(t *testing.T) {
			bs, err := yaml.Marshal(test.project)
			if err != nil {
				t.Fatal(err)
			}

			if string(bs) != test.expected {
				t.Fatalf("Expected %v but got %v", test.expected, string(bs))
			}
		})
	}
}

func TestUnmarshalProject(t *testing.T) {
	tests := []struct {
		note     string
		input    string
		expected *Project
	}{
		{
			note: "no dependencies",
			input: `name: test_project
version: 0.0.1
source: src
`,
			expected: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
			},
		},
		{
			note: "file dependency with only simplified name & location",
			input: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo: file://dev/null
`,
			expected: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",

				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "foo",
						},
					},
				},
			},
		},
		{
			note: "file dependency with only name & location",
			input: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo: 
        location: file://dev/null
`,
			expected: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",

				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "foo",
						},
					},
				},
			},
		},
		{
			note: "file dependency with no namespace",
			input: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo:
        location: file://dev/null
        namespace: false
`,
			expected: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "",
						},
					},
				},
			},
		},
		{
			note: "file dependency with named namespace",
			input: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo:
        location: file://dev/null
        namespace: bar
`,
			expected: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "file://dev/null",
							Namespace: "bar",
						},
					},
				},
			},
		},
		{
			note: "git dependency with only name & location",
			input: `name: test_project
version: 0.0.1
source: src
dependencies:
    foo: git+https://example.com/my/repo
`,
			expected: &Project{
				Name:      "test_project",
				Version:   "0.0.1",
				SourceDir: "src",
				Dependencies: Dependencies{
					"foo": Dependency{
						Name: "foo",
						DependencyInfo: DependencyInfo{
							Location:  "git+https://example.com/my/repo",
							Namespace: "foo",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.note, func(t *testing.T) {
			var project Project
			err := yaml.Unmarshal([]byte(test.input), &project)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(&project, test.expected) {
				t.Fatalf("Expected %v but got %v", test.expected, project)
			}
		})
	}
}

// proj
// +-- dep_a (no opa.project)
// +-- dep_b (opa.project, no source)
// |   +-- dep_b1 (no opa.project)
// |   +-- dep_b2 (opa.project, source)
// +-- dep_c (opa.project, source)
// .   +-- dep_c1 (opa.project, no source)
// .   +-- dep_c2 (opa.project, source)
func TestReadProjectFromFile(t *testing.T) {
	files := map[string]string{
		"opa.project": `name: proj
source: src
dependencies:
  dep_a: file://dep_a
  dep_b: file://dep_b
  dep_c: file://dep_c
`,
		"src/policy.rego":                     `package test`,
		".opa/dependencies/dep_a/policy.rego": `package dep_a`,
		".opa/dependencies/dep_b/opa.project": `name: dep_b
dependencies:
  dep_b1: file://dep_b1
  dep_b2: file://dep_b2`,
		".opa/dependencies/dep_b/.opa/dependencies/dep_b1/policy.rego": `package dep_b1`,
		".opa/dependencies/dep_b/.opa/dependencies/dep_b2/opa.project": `name: dep_b2
source: foo`,
		".opa/dependencies/dep_b/.opa/dependencies/dep_b2/foobar/policy.rego": `package dep_b2`,
		".opa/dependencies/dep_c/opa.project": `name: dep_c
source: bar
dependencies:
  dep_c1: file://dep_c1
  dep_c2: file://dep_c2`,
		".opa/dependencies/dep_c/bar/policy.rego":                      `package dep_c`,
		".opa/dependencies/dep_c/.opa/dependencies/dep_c1/opa.project": `name: dep_c1`,
		".opa/dependencies/dep_c/.opa/dependencies/dep_c1/policy.rego": `package dep_c1`,
		".opa/dependencies/dep_c/.opa/dependencies/dep_c2/opa.project": `name: dep_c2
source: baz`,
		".opa/dependencies/dep_c/.opa/dependencies/dep_c2/baz/policy.rego": `package dep_c2`,
	}
	err := withTempFiles(files, func(path string) {
		fmt.Println(path)
		project, err := ReadProjectFromFile(path, false)
		if err != nil {
			t.Fatal(err)
		}
		if err := project.Load(); err != nil {
			t.Fatal(err)
		}
		dataLocations, err := project.DataLocations()
		if err != nil {
			t.Fatal(err)
		}

		_ = project.PrintTree(os.Stdout)

		expected := []string{
			filepath.Join(path, "src"),
			filepath.Join(path, ".opa/dependencies/dep_a"),
			filepath.Join(path, ".opa/dependencies/dep_b"),
			filepath.Join(path, ".opa/dependencies/dep_b/.opa/dependencies/dep_b1"),
			filepath.Join(path, ".opa/dependencies/dep_b/.opa/dependencies/dep_b2/foo"),
			filepath.Join(path, ".opa/dependencies/dep_c/bar"),
			filepath.Join(path, ".opa/dependencies/dep_c/.opa/dependencies/dep_c1"),
			filepath.Join(path, ".opa/dependencies/dep_c/.opa/dependencies/dep_c2/baz"),
		}

		if len(dataLocations) != len(expected) {
			t.Fatalf("Expected\n\n%v\n\nbut got\n\n%v", expected, dataLocations)
		}
		for _, e := range expected {
			found := false
			for _, d := range dataLocations {
				if e == d {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Expected\n\n%v\n\nbut got\n\n%v", expected, dataLocations)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func withTempFiles(files map[string]string, f func(string)) error {
	root, err := os.MkdirTemp("", "test-")
	if err != nil {
		return err
	}

	cleanup := func() {
		_ = os.RemoveAll(root)
	}
	defer cleanup()

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0644); err != nil {
			return err
		}
	}

	f(root)
	return nil
}
