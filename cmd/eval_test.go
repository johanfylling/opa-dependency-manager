package cmd

import (
	"bytes"
	"github.com/johanfylling/odm/printer"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEvalProjects(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(file)

	tests := []struct {
		name           string
		projectDir     string
		query          string
		expectedOutput string
	}{
		{
			name:           "Empty project",
			projectDir:     filepath.Join(rootDir, "testdata", "projects", "empty"),
			query:          "data",
			expectedOutput: `{}`,
		},
		{
			name:       "Empty project, call in query",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "empty"),
			query:      "x := 1 + 2",
			expectedOutput: `{
  "x": 3
}`,
		},
		{
			name:       "Project with source, no dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "no-dependencies"),
			query:      "x := data",
			expectedOutput: `{
  "x": {
    "test": {
      "allow": true
    }
  }
}`,
		},
		{
			name:       "Project with multiple source dirs, no dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "source-list"),
			query:      "x := data",
			expectedOutput: `{
  "x": {
    "do": {
      "re": {
        "mi": "fa"
      }
    },
    "foo": {
      "bar": "baz"
    },
    "test": {
      "allow": true
    }
  }
}`,
		},
		{
			name:       "Project with local dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "local-dependencies"),
			query:      "x := data",
			expectedOutput: `{
  "x": {
    "no_deps": {
      "test": {
        "allow": true
      }
    }
  }
}`,
		},
		{
			name:       "Project with transitive dependencies",
			projectDir: filepath.Join(rootDir, "testdata", "projects", "transitive-dependencies"),
			query:      "x := data",
			expectedOutput: `{
  "x": {
    "bar": {
      "no_deps": {
        "test": {
          "allow": true
        }
      }
    },
    "foo": {
      "no_deps": {
        "test": {
          "allow": true
        }
      }
    },
    "no_deps": {
      "test": {
        "allow": true
      }
    }
  }
}`,
		},
	}

	for _, tc := range tests {
		//goland:noinspection GoDeferInLoop
		defer cleanup(tc.projectDir)

		t.Run(tc.name, func(t *testing.T) {
			output := bytes.Buffer{}
			printer.PrintWriter = &output
			args := []string{tc.query, "--format", "bindings"}
			if err := doUpdate(tc.projectDir); err != nil {
				t.Fatal(err)
			}
			if err := doEval(tc.projectDir, args); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), tc.expectedOutput) {
				t.Fatalf("expected output:\n\n%s\n\ngot:\n\n%s", tc.expectedOutput, output.String())
			}
		})
	}
}
