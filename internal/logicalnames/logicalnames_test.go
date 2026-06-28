// code-from-spec: SPEC/golang/tests/utils/logical_names@3qZJRCLjh3tsGDNQYNPEbWeTzbs
package logicalnames_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

func testStringPtr(s string) *string {
	return &s
}

func testCreateFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath(path), 0755); err != nil {
		t.Fatalf("testCreateFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testCreateFile write: %v", err)
	}
}

func filepath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func TestLogicalNameParse_Spec(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantType      logicalnames.NodeType
		wantName      string
		wantQualifier *string
		wantPath      string
		wantParent    *string
		wantErr       error
	}{
		{
			name:          "SPEC alone",
			input:         "SPEC",
			wantType:      logicalnames.NodeTypeSpec,
			wantName:      "SPEC",
			wantQualifier: nil,
			wantPath:      "code-from-spec/_node.md",
			wantParent:    nil,
		},
		{
			name:          "SPEC with single segment",
			input:         "SPEC/domain",
			wantType:      logicalnames.NodeTypeSpec,
			wantName:      "SPEC/domain",
			wantQualifier: nil,
			wantPath:      "code-from-spec/domain/_node.md",
			wantParent:    testStringPtr("SPEC"),
		},
		{
			name:          "SPEC with nested path",
			input:         "SPEC/payments/fees/calculation",
			wantType:      logicalnames.NodeTypeSpec,
			wantName:      "SPEC/payments/fees/calculation",
			wantQualifier: nil,
			wantPath:      "code-from-spec/payments/fees/calculation/_node.md",
			wantParent:    testStringPtr("SPEC/payments/fees"),
		},
		{
			name:          "SPEC with qualifier",
			input:         "SPEC/x/y(interface)",
			wantType:      logicalnames.NodeTypeSpec,
			wantName:      "SPEC/x/y",
			wantQualifier: testStringPtr("interface"),
			wantPath:      "code-from-spec/x/y/_node.md",
			wantParent:    testStringPtr("SPEC/x"),
		},
		{
			name:          "SPEC with qualifier root level",
			input:         "SPEC(context)",
			wantType:      logicalnames.NodeTypeSpec,
			wantName:      "SPEC",
			wantQualifier: testStringPtr("context"),
			wantPath:      "code-from-spec/_node.md",
			wantParent:    nil,
		},
		{
			name:          "SPEC with qualifier parent computed from unqualified name",
			input:         "SPEC/domain/config(interface)",
			wantType:      logicalnames.NodeTypeSpec,
			wantName:      "SPEC/domain/config",
			wantQualifier: testStringPtr("interface"),
			wantPath:      "code-from-spec/domain/config/_node.md",
			wantParent:    testStringPtr("SPEC/domain"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameParse(tc.input)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Type != tc.wantType {
				t.Errorf("Type: expected %v, got %v", tc.wantType, got.Type)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name: expected %q, got %q", tc.wantName, got.Name)
			}
			if tc.wantQualifier == nil {
				if got.Qualifier != nil {
					t.Errorf("Qualifier: expected nil, got %q", *got.Qualifier)
				}
			} else {
				if got.Qualifier == nil {
					t.Errorf("Qualifier: expected %q, got nil", *tc.wantQualifier)
				} else if *got.Qualifier != *tc.wantQualifier {
					t.Errorf("Qualifier: expected %q, got %q", *tc.wantQualifier, *got.Qualifier)
				}
			}
			if got.Path != tc.wantPath {
				t.Errorf("Path: expected %q, got %q", tc.wantPath, got.Path)
			}
			if tc.wantParent == nil {
				if got.Parent != nil {
					t.Errorf("Parent: expected nil, got %q", *got.Parent)
				}
			} else {
				if got.Parent == nil {
					t.Errorf("Parent: expected %q, got nil", *tc.wantParent)
				} else if *got.Parent != *tc.wantParent {
					t.Errorf("Parent: expected %q, got %q", *tc.wantParent, *got.Parent)
				}
			}
		})
	}
}

func TestLogicalNameParse_External(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantPath   string
		wantParent *string
	}{
		{
			name:       "Simple external path",
			input:      "EXTERNAL/proto/v1/api.proto",
			wantName:   "EXTERNAL/proto/v1/api.proto",
			wantPath:   "proto/v1/api.proto",
			wantParent: nil,
		},
		{
			name:       "Root-level external file",
			input:      "EXTERNAL/docker-compose.yaml",
			wantName:   "EXTERNAL/docker-compose.yaml",
			wantPath:   "docker-compose.yaml",
			wantParent: nil,
		},
		{
			name:       "Deeply nested external path",
			input:      "EXTERNAL/a/b/c/d/schema.proto",
			wantName:   "EXTERNAL/a/b/c/d/schema.proto",
			wantPath:   "a/b/c/d/schema.proto",
			wantParent: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameParse(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Type != logicalnames.NodeTypeExternal {
				t.Errorf("Type: expected NodeTypeExternal, got %v", got.Type)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name: expected %q, got %q", tc.wantName, got.Name)
			}
			if got.Qualifier != nil {
				t.Errorf("Qualifier: expected nil, got %q", *got.Qualifier)
			}
			if got.Path != tc.wantPath {
				t.Errorf("Path: expected %q, got %q", tc.wantPath, got.Path)
			}
			if tc.wantParent == nil {
				if got.Parent != nil {
					t.Errorf("Parent: expected nil, got %q", *got.Parent)
				}
			} else {
				if got.Parent == nil {
					t.Errorf("Parent: expected %q, got nil", *tc.wantParent)
				} else if *got.Parent != *tc.wantParent {
					t.Errorf("Parent: expected %q, got %q", *tc.wantParent, *got.Parent)
				}
			}
		})
	}
}

func TestLogicalNameParse_Artifact(t *testing.T) {
	t.Run("Simple artifact", func(t *testing.T) {
		tmp := t.TempDir()
		testChdir(t, tmp)
		testCreateFile(t, "code-from-spec/extraction/proto/_node.md",
			"---\noutput: internal/extraction/proto.go\n---\n# SPEC/extraction/proto\n")

		got, err := logicalnames.LogicalNameParse("ARTIFACT/extraction/proto")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Type != logicalnames.NodeTypeArtifact {
			t.Errorf("Type: expected NodeTypeArtifact, got %v", got.Type)
		}
		if got.Name != "ARTIFACT/extraction/proto" {
			t.Errorf("Name: expected %q, got %q", "ARTIFACT/extraction/proto", got.Name)
		}
		if got.Qualifier != nil {
			t.Errorf("Qualifier: expected nil, got %q", *got.Qualifier)
		}
		if got.Path != "internal/extraction/proto.go" {
			t.Errorf("Path: expected %q, got %q", "internal/extraction/proto.go", got.Path)
		}
		if got.Parent == nil {
			t.Errorf("Parent: expected %q, got nil", "SPEC/extraction/proto")
		} else if *got.Parent != "SPEC/extraction/proto" {
			t.Errorf("Parent: expected %q, got %q", "SPEC/extraction/proto", *got.Parent)
		}
	})

	t.Run("Artifact with nested generator", func(t *testing.T) {
		tmp := t.TempDir()
		testChdir(t, tmp)
		testCreateFile(t, "code-from-spec/payments/fees/calculation/_node.md",
			"---\noutput: internal/fees/calculation.go\n---\n# SPEC/payments/fees/calculation\n")

		got, err := logicalnames.LogicalNameParse("ARTIFACT/payments/fees/calculation")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Type != logicalnames.NodeTypeArtifact {
			t.Errorf("Type: expected NodeTypeArtifact, got %v", got.Type)
		}
		if got.Name != "ARTIFACT/payments/fees/calculation" {
			t.Errorf("Name: expected %q, got %q", "ARTIFACT/payments/fees/calculation", got.Name)
		}
		if got.Qualifier != nil {
			t.Errorf("Qualifier: expected nil, got %q", *got.Qualifier)
		}
		if got.Path != "internal/fees/calculation.go" {
			t.Errorf("Path: expected %q, got %q", "internal/fees/calculation.go", got.Path)
		}
		if got.Parent == nil {
			t.Errorf("Parent: expected %q, got nil", "SPEC/payments/fees/calculation")
		} else if *got.Parent != "SPEC/payments/fees/calculation" {
			t.Errorf("Parent: expected %q, got %q", "SPEC/payments/fees/calculation", *got.Parent)
		}
	})

	t.Run("Artifact generator has no output", func(t *testing.T) {
		tmp := t.TempDir()
		testChdir(t, tmp)
		testCreateFile(t, "code-from-spec/docs/overview/_node.md",
			"---\n---\n# SPEC/docs/overview\n")

		_, err := logicalnames.LogicalNameParse("ARTIFACT/docs/overview")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, logicalnames.ErrNoOutput) {
			t.Fatalf("expected ErrNoOutput, got %v", err)
		}
	})

	t.Run("Artifact generator does not exist on disk", func(t *testing.T) {
		tmp := t.TempDir()
		testChdir(t, tmp)

		_, err := logicalnames.LogicalNameParse("ARTIFACT/nonexistent/node")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestLogicalNameParse_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "Unrecognized prefix",
			input:   "UNKNOWN/something",
			wantErr: logicalnames.ErrUnrecognizedPrefix,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: logicalnames.ErrUnrecognizedPrefix,
		},
		{
			name:    "ROOT prefix",
			input:   "ROOT/x",
			wantErr: logicalnames.ErrUnrecognizedPrefix,
		},
		{
			name:    "SPEC/ with empty relative path",
			input:   "SPEC/",
			wantErr: logicalnames.ErrInvalidName,
		},
		{
			name:    "ARTIFACT/ with empty relative path",
			input:   "ARTIFACT/",
			wantErr: logicalnames.ErrInvalidName,
		},
		{
			name:    "EXTERNAL/ with empty relative path",
			input:   "EXTERNAL/",
			wantErr: logicalnames.ErrInvalidName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := logicalnames.LogicalNameParse(tc.input)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.wantErr)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestLogicalNameFromPath(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantType   logicalnames.NodeType
		wantName   string
		wantPath   string
		wantParent *string
		wantErr    error
	}{
		{
			name:       "Root node",
			input:      "code-from-spec/_node.md",
			wantType:   logicalnames.NodeTypeSpec,
			wantName:   "SPEC",
			wantPath:   "code-from-spec/_node.md",
			wantParent: nil,
		},
		{
			name:       "Nested node",
			input:      "code-from-spec/x/y/_node.md",
			wantType:   logicalnames.NodeTypeSpec,
			wantName:   "SPEC/x/y",
			wantPath:   "code-from-spec/x/y/_node.md",
			wantParent: testStringPtr("SPEC/x"),
		},
		{
			name:       "Deeply nested node",
			input:      "code-from-spec/a/b/c/d/_node.md",
			wantType:   logicalnames.NodeTypeSpec,
			wantName:   "SPEC/a/b/c/d",
			wantPath:   "code-from-spec/a/b/c/d/_node.md",
			wantParent: testStringPtr("SPEC/a/b/c"),
		},
		{
			name:    "Rejects non-spec path",
			input:   "internal/config/config.go",
			wantErr: logicalnames.ErrInvalidPath,
		},
		{
			name:    "Rejects path without _node.md",
			input:   "code-from-spec/x/y/output.md",
			wantErr: logicalnames.ErrInvalidPath,
		},
		{
			name:    "Rejects path not starting with code-from-spec/",
			input:   "other/x/_node.md",
			wantErr: logicalnames.ErrInvalidPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameFromPath(pathutils.PathCfs{Value: tc.input})
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Type != tc.wantType {
				t.Errorf("Type: expected %v, got %v", tc.wantType, got.Type)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name: expected %q, got %q", tc.wantName, got.Name)
			}
			if got.Qualifier != nil {
				t.Errorf("Qualifier: expected nil, got %q", *got.Qualifier)
			}
			if got.Path != tc.wantPath {
				t.Errorf("Path: expected %q, got %q", tc.wantPath, got.Path)
			}
			if tc.wantParent == nil {
				if got.Parent != nil {
					t.Errorf("Parent: expected nil, got %q", *got.Parent)
				}
			} else {
				if got.Parent == nil {
					t.Errorf("Parent: expected %q, got nil", *tc.wantParent)
				} else if *got.Parent != *tc.wantParent {
					t.Errorf("Parent: expected %q, got %q", *tc.wantParent, *got.Parent)
				}
			}
		})
	}
}
