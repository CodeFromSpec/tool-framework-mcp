// code-from-spec: SPEC/golang/tests/parsing/logical_names@byLjJJDJLxgPo3wbYG1sZU8IUyo
package parsing_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
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

func ptrString(s string) *string {
	return &s
}

func TestCfsReferenceFromName_SpecType(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantErr        error
		wantNodeType   parsing.CfsNodeType
		wantLogical    string
		wantQualifier  *string
		wantPath       string
		wantParentName *string
	}{
		{
			name:    "bare SPEC is invalid",
			input:   "SPEC",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "bare ARTIFACT is invalid",
			input:   "ARTIFACT",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "bare EXTERNAL is invalid",
			input:   "EXTERNAL",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:           "SPEC root node single segment",
			input:          "SPEC/domain",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/domain/_node.md",
			wantParentName: nil,
		},
		{
			name:           "SPEC with nested path",
			input:          "SPEC/payments/fees/calculation",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/payments/fees/calculation",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/payments/fees/calculation/_node.md",
			wantParentName: ptrString("SPEC/payments/fees"),
		},
		{
			name:           "SPEC with qualifier",
			input:          "SPEC/x/y(interface)",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/x/y",
			wantQualifier:  ptrString("interface"),
			wantPath:       "code-from-spec/x/y/_node.md",
			wantParentName: ptrString("SPEC/x"),
		},
		{
			name:           "SPEC with qualifier root level",
			input:          "SPEC/domain(context)",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain",
			wantQualifier:  ptrString("context"),
			wantPath:       "code-from-spec/domain/_node.md",
			wantParentName: nil,
		},
		{
			name:           "SPEC with qualifier parent computed from unqualified name",
			input:          "SPEC/domain/config(interface)",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain/config",
			wantQualifier:  ptrString("interface"),
			wantPath:       "code-from-spec/domain/config/_node.md",
			wantParentName: ptrString("SPEC/domain"),
		},
		{
			name:    "SPEC/ with empty relative path",
			input:   "SPEC/",
			wantErr: parsing.ErrInvalidName,
		},
		{
			name:    "SPEC name with trailing slash",
			input:   "SPEC/a/b/",
			wantErr: parsing.ErrInvalidName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parsing.CfsReferenceFromName(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.NodeType != tc.wantNodeType {
				t.Errorf("NodeType: got %v, want %v", ref.NodeType, tc.wantNodeType)
			}
			if ref.LogicalName != tc.wantLogical {
				t.Errorf("LogicalName: got %q, want %q", ref.LogicalName, tc.wantLogical)
			}
			if tc.wantQualifier == nil {
				if ref.Qualifier != nil {
					t.Errorf("Qualifier: got %q, want nil", *ref.Qualifier)
				}
			} else {
				if ref.Qualifier == nil {
					t.Errorf("Qualifier: got nil, want %q", *tc.wantQualifier)
				} else if *ref.Qualifier != *tc.wantQualifier {
					t.Errorf("Qualifier: got %q, want %q", *ref.Qualifier, *tc.wantQualifier)
				}
			}
			if ref.Path != tc.wantPath {
				t.Errorf("Path: got %q, want %q", ref.Path, tc.wantPath)
			}
			if tc.wantParentName == nil {
				if ref.ParentName != nil {
					t.Errorf("ParentName: got %q, want nil", *ref.ParentName)
				}
			} else {
				if ref.ParentName == nil {
					t.Errorf("ParentName: got nil, want %q", *tc.wantParentName)
				} else if *ref.ParentName != *tc.wantParentName {
					t.Errorf("ParentName: got %q, want %q", *ref.ParentName, *tc.wantParentName)
				}
			}
		})
	}
}

func TestCfsReferenceFromName_ExternalType(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantNodeType   parsing.CfsNodeType
		wantLogical    string
		wantQualifier  *string
		wantPath       string
		wantParentName *string
	}{
		{
			name:           "simple external path",
			input:          "EXTERNAL/proto/v1/api.proto",
			wantNodeType:   parsing.CfsNodeTypeExternal,
			wantLogical:    "EXTERNAL/proto/v1/api.proto",
			wantQualifier:  nil,
			wantPath:       "proto/v1/api.proto",
			wantParentName: nil,
		},
		{
			name:           "root-level external file",
			input:          "EXTERNAL/docker-compose.yaml",
			wantNodeType:   parsing.CfsNodeTypeExternal,
			wantLogical:    "EXTERNAL/docker-compose.yaml",
			wantQualifier:  nil,
			wantPath:       "docker-compose.yaml",
			wantParentName: nil,
		},
		{
			name:           "deeply nested external path",
			input:          "EXTERNAL/a/b/c/d/schema.proto",
			wantNodeType:   parsing.CfsNodeTypeExternal,
			wantLogical:    "EXTERNAL/a/b/c/d/schema.proto",
			wantQualifier:  nil,
			wantPath:       "a/b/c/d/schema.proto",
			wantParentName: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parsing.CfsReferenceFromName(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.NodeType != tc.wantNodeType {
				t.Errorf("NodeType: got %v, want %v", ref.NodeType, tc.wantNodeType)
			}
			if ref.LogicalName != tc.wantLogical {
				t.Errorf("LogicalName: got %q, want %q", ref.LogicalName, tc.wantLogical)
			}
			if ref.Qualifier != nil {
				t.Errorf("Qualifier: got %q, want nil", *ref.Qualifier)
			}
			if ref.Path != tc.wantPath {
				t.Errorf("Path: got %q, want %q", ref.Path, tc.wantPath)
			}
			if ref.ParentName != nil {
				t.Errorf("ParentName: got %q, want nil", *ref.ParentName)
			}
		})
	}
}

func TestCfsReferenceFromName_ExternalErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "EXTERNAL/ with empty relative path",
			input:   "EXTERNAL/",
			wantErr: parsing.ErrInvalidName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsing.CfsReferenceFromName(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCfsReferenceFromName_ArtifactType(t *testing.T) {
	t.Run("simple artifact", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		dir := filepath.Join(tmpDir, "code-from-spec", "extraction", "proto")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		content := "---\noutput: internal/extraction/proto.go\n---\n# SPEC/extraction/proto\n"
		if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		ref, err := parsing.CfsReferenceFromName("ARTIFACT/extraction/proto")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeArtifact {
			t.Errorf("NodeType: got %v, want CfsNodeTypeArtifact", ref.NodeType)
		}
		if ref.LogicalName != "ARTIFACT/extraction/proto" {
			t.Errorf("LogicalName: got %q, want %q", ref.LogicalName, "ARTIFACT/extraction/proto")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier: got %q, want nil", *ref.Qualifier)
		}
		if ref.Path != "internal/extraction/proto.go" {
			t.Errorf("Path: got %q, want %q", ref.Path, "internal/extraction/proto.go")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName: got nil, want %q", "SPEC/extraction/proto")
		} else if *ref.ParentName != "SPEC/extraction/proto" {
			t.Errorf("ParentName: got %q, want %q", *ref.ParentName, "SPEC/extraction/proto")
		}
	})

	t.Run("artifact with nested generator", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		dir := filepath.Join(tmpDir, "code-from-spec", "payments", "fees", "calculation")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		content := "---\noutput: internal/fees/calculation.go\n---\n# SPEC/payments/fees/calculation\n"
		if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		ref, err := parsing.CfsReferenceFromName("ARTIFACT/payments/fees/calculation")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeArtifact {
			t.Errorf("NodeType: got %v, want CfsNodeTypeArtifact", ref.NodeType)
		}
		if ref.LogicalName != "ARTIFACT/payments/fees/calculation" {
			t.Errorf("LogicalName: got %q, want %q", ref.LogicalName, "ARTIFACT/payments/fees/calculation")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier: got %q, want nil", *ref.Qualifier)
		}
		if ref.Path != "internal/fees/calculation.go" {
			t.Errorf("Path: got %q, want %q", ref.Path, "internal/fees/calculation.go")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName: got nil, want %q", "SPEC/payments/fees/calculation")
		} else if *ref.ParentName != "SPEC/payments/fees/calculation" {
			t.Errorf("ParentName: got %q, want %q", *ref.ParentName, "SPEC/payments/fees/calculation")
		}
	})

	t.Run("artifact generator has no output", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		dir := filepath.Join(tmpDir, "code-from-spec", "docs", "overview")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		content := "---\n---\n# SPEC/docs/overview\n"
		if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		_, err := parsing.CfsReferenceFromName("ARTIFACT/docs/overview")
		if !errors.Is(err, parsing.ErrNoOutput) {
			t.Fatalf("expected ErrNoOutput, got %v", err)
		}
	})

	t.Run("artifact generator does not exist on disk", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		_, err := parsing.CfsReferenceFromName("ARTIFACT/nonexistent/node")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("ARTIFACT/ with empty relative path", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("ARTIFACT/")
		if !errors.Is(err, parsing.ErrInvalidName) {
			t.Fatalf("expected ErrInvalidName, got %v", err)
		}
	})
}

func TestCfsReferenceFromName_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "unrecognized prefix",
			input:   "UNKNOWN/something",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "ROOT prefix",
			input:   "ROOT/x",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsing.CfsReferenceFromName(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCfsReferenceFromPath(t *testing.T) {
	tests := []struct {
		name           string
		input          oslayer.CfsPath
		wantErr        error
		wantNodeType   parsing.CfsNodeType
		wantLogical    string
		wantQualifier  *string
		wantPath       string
		wantParentName *string
	}{
		{
			name:           "root node direct child of code-from-spec",
			input:          "code-from-spec/domain/_node.md",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/domain/_node.md",
			wantParentName: nil,
		},
		{
			name:           "nested node",
			input:          "code-from-spec/x/y/_node.md",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/x/y",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/x/y/_node.md",
			wantParentName: ptrString("SPEC/x"),
		},
		{
			name:           "deeply nested node",
			input:          "code-from-spec/a/b/c/d/_node.md",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/a/b/c/d",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/a/b/c/d/_node.md",
			wantParentName: ptrString("SPEC/a/b/c"),
		},
		{
			name:    "rejects bare code-from-spec/_node.md",
			input:   "code-from-spec/_node.md",
			wantErr: parsing.ErrInvalidPath,
		},
		{
			name:    "rejects non-spec path",
			input:   "internal/config/config.go",
			wantErr: parsing.ErrInvalidPath,
		},
		{
			name:    "rejects path without _node.md",
			input:   "code-from-spec/x/y/output.md",
			wantErr: parsing.ErrInvalidPath,
		},
		{
			name:    "rejects path not starting with code-from-spec/",
			input:   "other/x/_node.md",
			wantErr: parsing.ErrInvalidPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parsing.CfsReferenceFromPath(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.NodeType != tc.wantNodeType {
				t.Errorf("NodeType: got %v, want %v", ref.NodeType, tc.wantNodeType)
			}
			if ref.LogicalName != tc.wantLogical {
				t.Errorf("LogicalName: got %q, want %q", ref.LogicalName, tc.wantLogical)
			}
			if tc.wantQualifier == nil {
				if ref.Qualifier != nil {
					t.Errorf("Qualifier: got %q, want nil", *ref.Qualifier)
				}
			} else {
				if ref.Qualifier == nil {
					t.Errorf("Qualifier: got nil, want %q", *tc.wantQualifier)
				} else if *ref.Qualifier != *tc.wantQualifier {
					t.Errorf("Qualifier: got %q, want %q", *ref.Qualifier, *tc.wantQualifier)
				}
			}
			if ref.Path != tc.wantPath {
				t.Errorf("Path: got %q, want %q", ref.Path, tc.wantPath)
			}
			if tc.wantParentName == nil {
				if ref.ParentName != nil {
					t.Errorf("ParentName: got %q, want nil", *ref.ParentName)
				}
			} else {
				if ref.ParentName == nil {
					t.Errorf("ParentName: got nil, want %q", *tc.wantParentName)
				} else if *ref.ParentName != *tc.wantParentName {
					t.Errorf("ParentName: got %q, want %q", *ref.ParentName, *tc.wantParentName)
				}
			}
		})
	}
}
