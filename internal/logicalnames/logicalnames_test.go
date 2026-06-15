// code-from-spec: SPEC/golang/tests/utils/logical_names@kpD3x3JufFRCX2uF2cX8Xe5h3W8
package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func TestLogicalNameToPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPath    string
		wantErr     error
	}{
		{
			name:     "SPEC alone",
			input:    "SPEC",
			wantPath: "code-from-spec/_node.md",
		},
		{
			name:     "SPEC with path",
			input:    "SPEC/payments/processor",
			wantPath: "code-from-spec/payments/processor/_node.md",
		},
		{
			name:     "Strips qualifier before resolving",
			input:    "SPEC/x/y(interface)",
			wantPath: "code-from-spec/x/y/_node.md",
		},
		{
			name:    "Rejects ROOT reference",
			input:   "ROOT/x",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		{
			name:    "Rejects ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		{
			name:    "Rejects EXTERNAL reference",
			input:   "EXTERNAL/proto/api.proto",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		{
			name:    "Rejects unrecognized prefix",
			input:   "UNKNOWN/something",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		{
			name:    "Rejects empty string",
			input:   "",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameToPath(tc.input)
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
			if got == nil {
				t.Fatal("got nil PathCfs")
			}
			if got.Value != tc.wantPath {
				t.Fatalf("expected %q, got %q", tc.wantPath, got.Value)
			}
		})
	}
}

func TestLogicalNameFromPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "Root node",
			input: "code-from-spec/_node.md",
			want:  "SPEC",
		},
		{
			name:  "Nested node",
			input: "code-from-spec/x/y/_node.md",
			want:  "SPEC/x/y",
		},
		{
			name:    "Rejects non-node path",
			input:   "internal/config/config.go",
			wantErr: logicalnames.ErrInvalidPath,
		},
		{
			name:    "Rejects path without _node.md",
			input:   "code-from-spec/x/y/output.md",
			wantErr: logicalnames.ErrInvalidPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: tc.input})
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
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestLogicalNameGetParent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "SPEC/x parent is SPEC",
			input: "SPEC/domain",
			want:  "SPEC",
		},
		{
			name:  "SPEC/x/y parent is SPEC/x",
			input: "SPEC/domain/config",
			want:  "SPEC/domain",
		},
		{
			name:  "Strips qualifier before computing parent",
			input: "SPEC/domain/config(interface)",
			want:  "SPEC/domain",
		},
		{
			name:    "SPEC has no parent",
			input:   "SPEC",
			wantErr: logicalnames.ErrNoParent,
		},
		{
			name:    "Rejects ROOT reference",
			input:   "ROOT/domain",
			wantErr: logicalnames.ErrNotASpecReference,
		},
		{
			name:    "Rejects ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantErr: logicalnames.ErrNotASpecReference,
		},
		{
			name:    "Rejects EXTERNAL reference",
			input:   "EXTERNAL/x",
			wantErr: logicalnames.ErrNotASpecReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetParent(tc.input)
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
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestLogicalNameGetQualifier(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantQual  string
		wantOk    bool
	}{
		{
			name:     "Extracts qualifier from SPEC reference",
			input:    "SPEC/x/y(interface)",
			wantQual: "interface",
			wantOk:   true,
		},
		{
			name:   "ARTIFACT without qualifier returns absent",
			input:  "ARTIFACT/x/y",
			wantOk: false,
		},
		{
			name:   "EXTERNAL without qualifier returns absent",
			input:  "EXTERNAL/proto/api.proto",
			wantOk: false,
		},
		{
			name:   "Returns absent when no qualifier",
			input:  "SPEC/x/y",
			wantOk: false,
		},
		{
			name:   "Returns absent for SPEC alone",
			input:  "SPEC",
			wantOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := logicalnames.LogicalNameGetQualifier(tc.input)
			if ok != tc.wantOk {
				t.Fatalf("expected ok=%v, got ok=%v", tc.wantOk, ok)
			}
			if ok && got != tc.wantQual {
				t.Fatalf("expected qualifier %q, got %q", tc.wantQual, got)
			}
		})
	}
}

func TestLogicalNameStripQualifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Strips qualifier from SPEC reference",
			input: "SPEC/x/y(interface)",
			want:  "SPEC/x/y",
		},
		{
			name:  "ARTIFACT without qualifier returns unchanged",
			input: "ARTIFACT/x/y",
			want:  "ARTIFACT/x/y",
		},
		{
			name:  "EXTERNAL returns unchanged",
			input: "EXTERNAL/proto/api.proto",
			want:  "EXTERNAL/proto/api.proto",
		},
		{
			name:  "No qualifier returns unchanged",
			input: "SPEC/x/y",
			want:  "SPEC/x/y",
		},
		{
			name:  "SPEC alone returns unchanged",
			input: "SPEC",
			want:  "SPEC",
		},
		{
			name:  "Empty string returns unchanged",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameStripQualifier(tc.input)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestLogicalNameHasParent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "SPEC alone",
			input: "SPEC",
			want:  false,
		},
		{
			name:  "SPEC with path",
			input: "SPEC/domain/config",
			want:  true,
		},
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x",
			want:  false,
		},
		{
			name:  "EXTERNAL reference",
			input: "EXTERNAL/x",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLogicalNameHasQualifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "Without qualifier",
			input: "SPEC/x",
			want:  false,
		},
		{
			name:  "With qualifier",
			input: "SPEC/x(y)",
			want:  true,
		},
		{
			name:  "ARTIFACT without qualifier",
			input: "ARTIFACT/x",
			want:  false,
		},
		{
			name:  "EXTERNAL without qualifier",
			input: "EXTERNAL/x",
			want:  false,
		},
		{
			name:  "SPEC alone",
			input: "SPEC",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasQualifier(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLogicalNameIsArtifact(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x",
			want:  true,
		},
		{
			name:  "SPEC reference",
			input: "SPEC/x(y)",
			want:  false,
		},
		{
			name:  "EXTERNAL reference",
			input: "EXTERNAL/x",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsArtifact(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLogicalNameIsSpec(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "SPEC alone",
			input: "SPEC",
			want:  true,
		},
		{
			name:  "SPEC with path",
			input: "SPEC/x/y",
			want:  true,
		},
		{
			name:  "ROOT reference is not SPEC",
			input: "ROOT/x",
			want:  false,
		},
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x",
			want:  false,
		},
		{
			name:  "EXTERNAL reference",
			input: "EXTERNAL/x",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsSpec(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLogicalNameIsExternal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "EXTERNAL reference",
			input: "EXTERNAL/proto/api.proto",
			want:  true,
		},
		{
			name:  "SPEC reference",
			input: "SPEC/x",
			want:  false,
		},
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsExternal(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLogicalNameGetArtifactGenerator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "Simple artifact",
			input: "ARTIFACT/x",
			want:  "SPEC/x",
		},
		{
			name:  "Nested artifact",
			input: "ARTIFACT/x/y/z",
			want:  "SPEC/x/y/z",
		},
		{
			name:    "Rejects SPEC reference",
			input:   "SPEC/x(y)",
			wantErr: logicalnames.ErrNotAnArtifactReference,
		},
		{
			name:    "Rejects EXTERNAL reference",
			input:   "EXTERNAL/x",
			wantErr: logicalnames.ErrNotAnArtifactReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetArtifactGenerator(tc.input)
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
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestLogicalNameExternalToPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
		wantErr  error
	}{
		{
			name:     "Simple path",
			input:    "EXTERNAL/proto/v1/api.proto",
			wantPath: "proto/v1/api.proto",
		},
		{
			name:     "Root-level file",
			input:    "EXTERNAL/docker-compose.yaml",
			wantPath: "docker-compose.yaml",
		},
		{
			name:    "Rejects SPEC reference",
			input:   "SPEC/x",
			wantErr: logicalnames.ErrNotAnExternalReference,
		},
		{
			name:    "Rejects ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantErr: logicalnames.ErrNotAnExternalReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameExternalToPath(tc.input)
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
			if got == nil {
				t.Fatal("got nil PathCfs")
			}
			if got.Value != tc.wantPath {
				t.Fatalf("expected %q, got %q", tc.wantPath, got.Value)
			}
		})
	}
}
