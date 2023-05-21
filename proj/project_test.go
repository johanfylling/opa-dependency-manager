package proj

import (
	"gopkg.in/yaml.v3"
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
