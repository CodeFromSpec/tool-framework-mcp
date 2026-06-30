// code-from-spec: SPEC/golang/tests/parsing/logical_names@L406uVuyQ2UZGfyYdWbT78BHYko
package parsinglogicalnamestest

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

func ptrStr(s string) *string {
	return &s
}

func TestCfsReferenceFromName_Spec(t *testing.T) {
	t.Run("BareSpecIsInvalid", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("SPEC")
		if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
			t.Fatalf("expected ErrUnrecognizedPrefix, got %v", err)
		}
	})

	t.Run("BareArtifactIsInvalid", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("ARTIFACT")
		if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
			t.Fatalf("expected ErrUnrecognizedPrefix, got %v", err)
		}
	})

	t.Run("BareExternalIsInvalid", func(t *testing.T) {
		_, err := parsing.CfsReferenceFromName("EXTERNAL")
		if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
			t.Fatalf("expected ErrUnrecognizedPrefix, got %v", err)
		}
	})

	type specCase struct {
		name           string
		input          string
		wantNodeType   parsing.CfsNodeType
		wantLogical    string
		wantQualifier  *string
		wantPath       string
		wantParentName *string
	}

	cases := []specCase{
		{
			name:           "RootNode",
			input:          "SPEC/domain",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/domain/_node.md",
			wantParentName: nil,
		},
		{
			name:           "NestedPath",
			input:          "SPEC/payments/fees/calculation",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/payments/fees/calculation",
			wantQualifier:  nil,
			wantPath:       "code-from-spec/payments/fees/calculation/_node.md",
			wantParentName: ptrStr("SPEC/payments/fees"),
		},
		{
			name:           "WithQualifier",
			input:          "SPEC/x/y(interface)",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/x/y",
			wantQualifier:  ptrStr("interface"),
			wantPath:       "code-from-spec/x/y/_node.md",
			wantParentName: ptrStr("SPEC/x"),
		},
		{
			name:           "WithQualifierRootLevel",
			input:          "SPEC/domain(context)",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain",
			wantQualifier:  ptrStr("context"),
			wantPath:       "code-from-spec/domain/_node.md",
			wantParentName: nil,
		},
		{
			name:           "WithQualifierParentFromUnqualifiedName",
			input:          "SPEC/domain/config(interface)",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain/config",
			wantQualifier:  ptrStr("interface"),
			wantPath:       "code-from-spec/domain/config/_node.md",
			wantParentName: ptrStr("SPEC/domain"),
		},
	}

	for _, tc := range cases {
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

func TestCfsReferenceFromName_External(t *testing.T) {
	type extCase struct {
		name        string
		input       string
		wantLogical string
		wantPath    string
	}

	cases := []extCase{
		{
			name:        "SimpleExternalPath",
			input:       "EXTERNAL/proto/v1/api.proto",
			wantLogical: "EXTERNAL/proto/v1/api.proto",
			wantPath:    "proto/v1/api.proto",
		},
		{
			name:        "RootLevelExternalFile",
			input:       "EXTERNAL/docker-compose.yaml",
			wantLogical: "EXTERNAL/docker-compose.yaml",
			wantPath:    "docker-compose.yaml",
		},
		{
			name:        "DeeplyNestedExternalPath",
			input:       "EXTERNAL/a/b/c/d/schema.proto",
			wantLogical: "EXTERNAL/a/b/c/d/schema.proto",
			wantPath:    "a/b/c/d/schema.proto",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parsing.CfsReferenceFromName(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.NodeType != parsing.CfsNodeTypeExternal {
				t.Errorf("NodeType: got %v, want CfsNodeTypeExternal", ref.NodeType)
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

func TestCfsReferenceFromName_Artifact(t *testing.T) {
	t.Run("SimpleArtifact", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		dir := filepath.Join(tmpDir, "code-from-spec", "extraction", "proto")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		content := "---\noutput: internal/extraction/proto.go\n---\n# SPEC/extraction/proto\n"
		if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
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
			t.Error("ParentName: got nil, want \"SPEC/extraction/proto\"")
		} else if *ref.ParentName != "SPEC/extraction/proto" {
			t.Errorf("ParentName: got %q, want %q", *ref.ParentName, "SPEC/extraction/proto")
		}
	})

	t.Run("ArtifactWithNestedGenerator", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		dir := filepath.Join(tmpDir, "code-from-spec", "payments", "fees", "calculation")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		content := "---\noutput: internal/fees/calculation.go\n---\n# SPEC/payments/fees/calculation\n"
		if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
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
			t.Error("ParentName: got nil, want \"SPEC/payments/fees/calculation\"")
		} else if *ref.ParentName != "SPEC/payments/fees/calculation" {
			t.Errorf("ParentName: got %q, want %q", *ref.ParentName, "SPEC/payments/fees/calculation")
		}
	})

	t.Run("ArtifactGeneratorHasNoOutput", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		dir := filepath.Join(tmpDir, "code-from-spec", "docs", "overview")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		content := "---\n---\n# SPEC/docs/overview\n"
		if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := parsing.CfsReferenceFromName("ARTIFACT/docs/overview")
		if !errors.Is(err, parsing.ErrNoOutput) {
			t.Fatalf("expected ErrNoOutput, got %v", err)
		}
	})

	t.Run("ArtifactGeneratorDoesNotExistOnDisk", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)

		_, err := parsing.CfsReferenceFromName("ARTIFACT/nonexistent/node")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCfsReferenceFromName_Errors(t *testing.T) {
	type errCase struct {
		name    string
		input   string
		wantErr error
	}

	cases := []errCase{
		{
			name:    "UnrecognizedPrefix",
			input:   "UNKNOWN/something",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "EmptyString",
			input:   "",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "ROOTPrefix",
			input:   "ROOT/x",
			wantErr: parsing.ErrUnrecognizedPrefix,
		},
		{
			name:    "SPECWithEmptyRelativePath",
			input:   "SPEC/",
			wantErr: parsing.ErrInvalidName,
		},
		{
			name:    "SPECNameWithTrailingSlash",
			input:   "SPEC/a/b/",
			wantErr: parsing.ErrInvalidName,
		},
		{
			name:    "ARTIFACTWithEmptyRelativePath",
			input:   "ARTIFACT/",
			wantErr: parsing.ErrInvalidName,
		},
		{
			name:    "EXTERNALWithEmptyRelativePath",
			input:   "EXTERNAL/",
			wantErr: parsing.ErrInvalidName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsing.CfsReferenceFromName(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCfsReferenceFromPath(t *testing.T) {
	type pathCase struct {
		name           string
		input          oslayer.CfsPath
		wantErr        error
		wantNodeType   parsing.CfsNodeType
		wantLogical    string
		wantPath       string
		wantParentName *string
	}

	cases := []pathCase{
		{
			name:           "RootNode",
			input:          "code-from-spec/domain/_node.md",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/domain",
			wantPath:       "code-from-spec/domain/_node.md",
			wantParentName: nil,
		},
		{
			name:           "NestedNode",
			input:          "code-from-spec/x/y/_node.md",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/x/y",
			wantPath:       "code-from-spec/x/y/_node.md",
			wantParentName: ptrStr("SPEC/x"),
		},
		{
			name:           "DeeplyNestedNode",
			input:          "code-from-spec/a/b/c/d/_node.md",
			wantNodeType:   parsing.CfsNodeTypeSpec,
			wantLogical:    "SPEC/a/b/c/d",
			wantPath:       "code-from-spec/a/b/c/d/_node.md",
			wantParentName: ptrStr("SPEC/a/b/c"),
		},
		{
			name:    "RejectsBareCodeFromSpecNodeMd",
			input:   "code-from-spec/_node.md",
			wantErr: parsing.ErrInvalidPath,
		},
		{
			name:    "RejectsNonSpecPath",
			input:   "internal/config/config.go",
			wantErr: parsing.ErrInvalidPath,
		},
		{
			name:    "RejectsPathWithoutNodeMd",
			input:   "code-from-spec/x/y/output.md",
			wantErr: parsing.ErrInvalidPath,
		},
		{
			name:    "RejectsPathNotStartingWithCodeFromSpec",
			input:   "other/x/_node.md",
			wantErr: parsing.ErrInvalidPath,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parsing.CfsReferenceFromPath(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
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
			if ref.Qualifier != nil {
				t.Errorf("Qualifier: got %q, want nil", *ref.Qualifier)
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
