package cmd

import (
	"bytes"
	"github.com/johanfylling/odm/printer"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestTestProjects(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(file)

	tests := []struct {
		name           string
		projectDir     string
		expectedOutput string
	}{
		{
			name:           "Empty project",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "empty"),
			expectedOutput: ``,
		},
		{
			name:       "Project with source, no dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "no-dependencies"),
			expectedOutput: `%ROOT_DIR%/testdata/projects/no-dependencies/tst/tests.rego:
data.test.test_allow: PASS (%TIME%)
--------------------------------------------------------------------------------
PASS: 1/1
`,
		},
		{
			name:       "Project with multiple source dirs, no dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "source-list"),
			expectedOutput: `%ROOT_DIR%/testdata/projects/source-list/test/test.rego:
data.test.test_allow: PASS (%TIME%)
--------------------------------------------------------------------------------
PASS: 1/1
`,
		},
		{
			name:       "Project with local dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "local-dependencies"),
			expectedOutput: `%ROOT_DIR%/testdata/projects/local-dependencies/.opa/dependencies/4310d81b00f2b2cc64a7ecdccb1ec277c4e83c547c0369398aeb0f695a37e37e/tst/tests.rego:
data.no_deps.test.test_allow: PASS (%TIME%)
--------------------------------------------------------------------------------
PASS: 1/1
`,
		},
		{
			name:       "Project with transitive dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies"),
			expectedOutput: `%ROOT_DIR%/testdata/projects/transitive-dependencies/.opa/dependencies/4310d81b00f2b2cc64a7ecdccb1ec277c4e83c547c0369398aeb0f695a37e37e/tst/tests.rego:
data.no_deps.test.test_allow: PASS (%TIME%)

%ROOT_DIR%/testdata/projects/transitive-dependencies/.opa/dependencies/8f8992cd45d31e54855edaef07238cf7a5be7d8225250ea3dd5f821cec3efe2a/tst/tests.rego:
data.foo.no_deps.test.test_allow: PASS (%TIME%)

%ROOT_DIR%/testdata/projects/transitive-dependencies/.opa/dependencies/b5d22e423449cd9944eaccddc3b302e48d5f1c3634838160fe56f91a5c58407f/tst/tests.rego:
data.bar.no_deps.test.test_allow: PASS (%TIME%)
--------------------------------------------------------------------------------
PASS: 3/3`,
		},
	}

	r := regexp.MustCompile(`(FAIL|PASS) \(.*s\)`)

	for _, tc := range tests {
		//goland:noinspection GoDeferInLoop
		defer cleanup(tc.projectDir)

		t.Run(tc.name, func(t *testing.T) {
			output := bytes.Buffer{}
			printer.PrintWriter = &output
			args := []string{"-v"}
			if err := doUpdate(tc.projectDir); err != nil {
				t.Fatal(err)
			}
			if err := doTest(tc.projectDir, true, args); err != nil {
				t.Fatal(err)
			}
			actual := r.ReplaceAllString(output.String(), "$1 (%TIME%)")
			expected := strings.ReplaceAll(tc.expectedOutput, "%ROOT_DIR%", rootDir)
			if !strings.Contains(actual, expected) {
				t.Fatalf("expected output:\n\n%s\n\ngot:\n\n%s", expected, actual)
			}
		})
	}
}
