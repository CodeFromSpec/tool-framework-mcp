package parsinglogicalnamestest

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestCfsReferenceFromName_SpecType(t *testing.T) {
	t.Run("bare SPEC is invalid", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("SPEC")
		if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
			t.Fatalf("expected ErrUnrecognizedPrefix, got %v", err)
		}
	})

	t.Run("bare ARTIFACT is invalid", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("ARTIFACT")
		if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
			t.Fatalf("expected ErrUnrecognizedPrefix, got %v", err)
		}
	})

	t.Run("bare EXTERNAL is invalid", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("EXTERNAL")
		if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
			t.Fatalf("expected ErrUnrecognizedPrefix, got %v", err)
		}
	})

	t.Run("SPEC root node single segment", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromName("SPEC/domain")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/domain" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/domain")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "code-from-spec/domain/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/domain/_node.md")
		}
		if ref.ParentName != nil {
			t.Errorf("ParentName = %v, want nil", ref.ParentName)
		}
	})

	t.Run("SPEC with nested path", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromName("SPEC/payments/fees/calculation")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/payments/fees/calculation" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/payments/fees/calculation")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "code-from-spec/payments/fees/calculation/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/payments/fees/calculation/_node.md")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/payments/fees")
		} else if *ref.ParentName != "SPEC/payments/fees" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/payments/fees")
		}
	})

	t.Run("SPEC with qualifier", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromName("SPEC/x/y(interface)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/x/y" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/x/y")
		}
		if ref.Qualifier == nil {
			t.Errorf("Qualifier = nil, want pointer to %q", "interface")
		} else if *ref.Qualifier != "interface" {
			t.Errorf("Qualifier = %q, want %q", *ref.Qualifier, "interface")
		}
		if ref.Path != "code-from-spec/x/y/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/x/y/_node.md")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/x")
		} else if *ref.ParentName != "SPEC/x" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/x")
		}
	})

	t.Run("SPEC with qualifier root level", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromName("SPEC/domain(context)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/domain" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/domain")
		}
		if ref.Qualifier == nil {
			t.Errorf("Qualifier = nil, want pointer to %q", "context")
		} else if *ref.Qualifier != "context" {
			t.Errorf("Qualifier = %q, want %q", *ref.Qualifier, "context")
		}
		if ref.Path != "code-from-spec/domain/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/domain/_node.md")
		}
		if ref.ParentName != nil {
			t.Errorf("ParentName = %v, want nil", ref.ParentName)
		}
	})

	t.Run("SPEC with qualifier parent computed from unqualified name", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromName("SPEC/domain/config(interface)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/domain/config" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/domain/config")
		}
		if ref.Qualifier == nil {
			t.Errorf("Qualifier = nil, want pointer to %q", "interface")
		} else if *ref.Qualifier != "interface" {
			t.Errorf("Qualifier = %q, want %q", *ref.Qualifier, "interface")
		}
		if ref.Path != "code-from-spec/domain/config/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/domain/config/_node.md")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/domain")
		} else if *ref.ParentName != "SPEC/domain" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/domain")
		}
	})
}

func TestCfsReferenceFromName_ExternalType(t *testing.T) {
	cases := []struct {
		name            string
		input           string
		wantLogicalName string
		wantPath        string
	}{
		{
			name:            "simple external path",
			input:           "EXTERNAL/proto/v1/api.proto",
			wantLogicalName: "EXTERNAL/proto/v1/api.proto",
			wantPath:        "proto/v1/api.proto",
		},
		{
			name:            "root-level external file",
			input:           "EXTERNAL/docker-compose.yaml",
			wantLogicalName: "EXTERNAL/docker-compose.yaml",
			wantPath:        "docker-compose.yaml",
		},
		{
			name:            "deeply nested external path",
			input:           "EXTERNAL/a/b/c/d/schema.proto",
			wantLogicalName: "EXTERNAL/a/b/c/d/schema.proto",
			wantPath:        "a/b/c/d/schema.proto",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parsing.CfsReferenceFromName(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.NodeType != parsing.CfsNodeTypeExternal {
				t.Errorf("NodeType = %v, want CfsNodeTypeExternal", ref.NodeType)
			}
			if ref.LogicalName != tc.wantLogicalName {
				t.Errorf("LogicalName = %q, want %q", ref.LogicalName, tc.wantLogicalName)
			}
			if ref.Qualifier != nil {
				t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
			}
			if ref.Path != tc.wantPath {
				t.Errorf("Path = %q, want %q", ref.Path, tc.wantPath)
			}
			if ref.ParentName != nil {
				t.Errorf("ParentName = %v, want nil", ref.ParentName)
			}
		})
	}
}

func TestCfsReferenceFromName_ArtifactType(t *testing.T) {
	t.Run("simple artifact", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/extraction/proto")
		b.SetOutput("internal/extraction/proto.go")
		b.Write()

		ref, err := parsing.CfsReferenceFromName("ARTIFACT/extraction/proto")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeArtifact {
			t.Errorf("NodeType = %v, want CfsNodeTypeArtifact", ref.NodeType)
		}
		if ref.LogicalName != "ARTIFACT/extraction/proto" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "ARTIFACT/extraction/proto")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "internal/extraction/proto.go" {
			t.Errorf("Path = %q, want %q", ref.Path, "internal/extraction/proto.go")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/extraction/proto")
		} else if *ref.ParentName != "SPEC/extraction/proto" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/extraction/proto")
		}
	})

	t.Run("artifact with nested generator", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/payments/fees/calculation")
		b.SetOutput("internal/fees/calculation.go")
		b.Write()

		ref, err := parsing.CfsReferenceFromName("ARTIFACT/payments/fees/calculation")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeArtifact {
			t.Errorf("NodeType = %v, want CfsNodeTypeArtifact", ref.NodeType)
		}
		if ref.LogicalName != "ARTIFACT/payments/fees/calculation" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "ARTIFACT/payments/fees/calculation")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "internal/fees/calculation.go" {
			t.Errorf("Path = %q, want %q", ref.Path, "internal/fees/calculation.go")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/payments/fees/calculation")
		} else if *ref.ParentName != "SPEC/payments/fees/calculation" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/payments/fees/calculation")
		}
	})

	t.Run("artifact generator has no output", func(t *testing.T) {
		testutils.Chdir(t)

		testutils.WriteRawNode(t, "SPEC/docs/overview", "---\n---\n# SPEC/docs/overview\n")

		_, err := parsing.CfsReferenceFromName("ARTIFACT/docs/overview")
		if !errors.Is(err, parsing.ErrNoOutput) {
			t.Fatalf("expected ErrNoOutput, got %v", err)
		}
	})

	t.Run("artifact generator does not exist on disk", func(t *testing.T) {
		testutils.Chdir(t)

		_, err := parsing.CfsReferenceFromName("ARTIFACT/nonexistent/node")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCfsReferenceFromName_Errors(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantError error
	}{
		{
			name:      "unrecognized prefix",
			input:     "UNKNOWN/something",
			wantError: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:      "ROOT prefix",
			input:     "ROOT/x",
			wantError: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:      "SPEC/ with empty relative path",
			input:     "SPEC/",
			wantError: parsing.ErrInvalidName,
		},
		{
			name:      "SPEC name with trailing slash",
			input:     "SPEC/a/b/",
			wantError: parsing.ErrInvalidName,
		},
		{
			name:      "ARTIFACT/ with empty relative path",
			input:     "ARTIFACT/",
			wantError: parsing.ErrInvalidName,
		},
		{
			name:      "EXTERNAL/ with empty relative path",
			input:     "EXTERNAL/",
			wantError: parsing.ErrInvalidName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsing.CfsReferenceFromName(tc.input)
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("expected %v, got %v", tc.wantError, err)
			}
		})
	}
}

func TestCfsReferenceFromPath(t *testing.T) {
	t.Run("root node direct child of code-from-spec", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("code-from-spec/domain/_node.md"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/domain" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/domain")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "code-from-spec/domain/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/domain/_node.md")
		}
		if ref.ParentName != nil {
			t.Errorf("ParentName = %v, want nil", ref.ParentName)
		}
	})

	t.Run("nested node", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("code-from-spec/x/y/_node.md"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/x/y" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/x/y")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "code-from-spec/x/y/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/x/y/_node.md")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/x")
		} else if *ref.ParentName != "SPEC/x" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/x")
		}
	})

	t.Run("deeply nested node", func(t *testing.T) {
		ref, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("code-from-spec/a/b/c/d/_node.md"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref.NodeType != parsing.CfsNodeTypeSpec {
			t.Errorf("NodeType = %v, want CfsNodeTypeSpec", ref.NodeType)
		}
		if ref.LogicalName != "SPEC/a/b/c/d" {
			t.Errorf("LogicalName = %q, want %q", ref.LogicalName, "SPEC/a/b/c/d")
		}
		if ref.Qualifier != nil {
			t.Errorf("Qualifier = %v, want nil", ref.Qualifier)
		}
		if ref.Path != "code-from-spec/a/b/c/d/_node.md" {
			t.Errorf("Path = %q, want %q", ref.Path, "code-from-spec/a/b/c/d/_node.md")
		}
		if ref.ParentName == nil {
			t.Errorf("ParentName = nil, want pointer to %q", "SPEC/a/b/c")
		} else if *ref.ParentName != "SPEC/a/b/c" {
			t.Errorf("ParentName = %q, want %q", *ref.ParentName, "SPEC/a/b/c")
		}
	})

	t.Run("rejects bare code-from-spec/_node.md", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("code-from-spec/_node.md"))
		if !errors.Is(err, parsing.ErrInvalidPath) {
			t.Fatalf("expected ErrInvalidPath, got %v", err)
		}
	})

	t.Run("rejects non-spec path", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("internal/config/config.go"))
		if !errors.Is(err, parsing.ErrInvalidPath) {
			t.Fatalf("expected ErrInvalidPath, got %v", err)
		}
	})

	t.Run("rejects path without _node.md", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("code-from-spec/x/y/output.md"))
		if !errors.Is(err, parsing.ErrInvalidPath) {
			t.Fatalf("expected ErrInvalidPath, got %v", err)
		}
	})

	t.Run("rejects path not starting with code-from-spec/", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromPath(oslayer.CfsPath("other/x/_node.md"))
		if !errors.Is(err, parsing.ErrInvalidPath) {
			t.Fatalf("expected ErrInvalidPath, got %v", err)
		}
	})
}
