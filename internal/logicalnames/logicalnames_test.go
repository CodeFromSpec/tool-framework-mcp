// code-from-spec: ROOT/golang/tests/utils/logical_names@bvRMAb-i7wY8_myi-3mXTrWmDyk
package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func TestLogicalNameToPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
		wantErr  error
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
				t.Fatal("expected non-nil PathCfs, got nil")
			}
			if got.Value != tc.wantPath {
				t.Errorf("expected %q, got %q", tc.wantPath, got.Value)
			}
		})
	}
}

func TestLogicalNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		{
			name:     "Root node",
			input:    "code-from-spec/_node.md",
			wantName: "SPEC",
		},
		{
			name:     "Nested node",
			input:    "code-from-spec/x/y/_node.md",
			wantName: "SPEC/x/y",
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
			if got != tc.wantName {
				t.Errorf("expected %q, got %q", tc.wantName, got)
			}
		})
	}
}

func TestLogicalNameGetParent(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantParent string
		wantErr    error
	}{
		{
			name:       "SPEC/x parent is SPEC",
			input:      "SPEC/domain",
			wantParent: "SPEC",
		},
		{
			name:       "SPEC/x/y parent is SPEC/x",
			input:      "SPEC/domain/config",
			wantParent: "SPEC/domain",
		},
		{
			name:       "Strips qualifier before computing parent",
			input:      "SPEC/domain/config(interface)",
			wantParent: "SPEC/domain",
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
			if got != tc.wantParent {
				t.Errorf("expected %q, got %q", tc.wantParent, got)
			}
		})
	}
}

func TestLogicalNameGetQualifier(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantQualifier string
		wantOk        bool
	}{
		{
			name:          "Extracts qualifier from SPEC reference",
			input:         "SPEC/x/y(interface)",
			wantQualifier: "interface",
			wantOk:        true,
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
			gotQualifier, gotOk := logicalnames.LogicalNameGetQualifier(tc.input)
			if gotOk != tc.wantOk {
				t.Errorf("expected ok=%v, got ok=%v", tc.wantOk, gotOk)
			}
			if tc.wantOk && gotQualifier != tc.wantQualifier {
				t.Errorf("expected qualifier %q, got %q", tc.wantQualifier, gotQualifier)
			}
		})
	}
}

func TestLogicalNameStripQualifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut string
	}{
		{
			name:    "Strips qualifier from SPEC reference",
			input:   "SPEC/x/y(interface)",
			wantOut: "SPEC/x/y",
		},
		{
			name:    "ARTIFACT without qualifier returns unchanged",
			input:   "ARTIFACT/x/y",
			wantOut: "ARTIFACT/x/y",
		},
		{
			name:    "EXTERNAL returns unchanged",
			input:   "EXTERNAL/proto/api.proto",
			wantOut: "EXTERNAL/proto/api.proto",
		},
		{
			name:    "No qualifier returns unchanged",
			input:   "SPEC/x/y",
			wantOut: "SPEC/x/y",
		},
		{
			name:    "SPEC alone returns unchanged",
			input:   "SPEC",
			wantOut: "SPEC",
		},
		{
			name:    "Empty string returns unchanged",
			input:   "",
			wantOut: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameStripQualifier(tc.input)
			if got != tc.wantOut {
				t.Errorf("expected %q, got %q", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameHasParent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		{
			name:    "SPEC alone",
			input:   "SPEC",
			wantOut: false,
		},
		{
			name:    "SPEC with path",
			input:   "SPEC/domain/config",
			wantOut: true,
		},
		{
			name:    "ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantOut: false,
		},
		{
			name:    "EXTERNAL reference",
			input:   "EXTERNAL/x",
			wantOut: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.wantOut {
				t.Errorf("expected %v, got %v", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameHasQualifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		{
			name:    "Without qualifier",
			input:   "SPEC/x",
			wantOut: false,
		},
		{
			name:    "With qualifier",
			input:   "SPEC/x(y)",
			wantOut: true,
		},
		{
			name:    "ARTIFACT without qualifier",
			input:   "ARTIFACT/x",
			wantOut: false,
		},
		{
			name:    "EXTERNAL without qualifier",
			input:   "EXTERNAL/x",
			wantOut: false,
		},
		{
			name:    "SPEC alone",
			input:   "SPEC",
			wantOut: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasQualifier(tc.input)
			if got != tc.wantOut {
				t.Errorf("expected %v, got %v", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameIsArtifact(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		{
			name:    "ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantOut: true,
		},
		{
			name:    "SPEC reference",
			input:   "SPEC/x(y)",
			wantOut: false,
		},
		{
			name:    "EXTERNAL reference",
			input:   "EXTERNAL/x",
			wantOut: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsArtifact(tc.input)
			if got != tc.wantOut {
				t.Errorf("expected %v, got %v", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameIsSpec(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		{
			name:    "SPEC alone",
			input:   "SPEC",
			wantOut: true,
		},
		{
			name:    "SPEC with path",
			input:   "SPEC/x/y",
			wantOut: true,
		},
		{
			name:    "ROOT reference not SPEC",
			input:   "ROOT/x",
			wantOut: false,
		},
		{
			name:    "ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantOut: false,
		},
		{
			name:    "EXTERNAL reference",
			input:   "EXTERNAL/x",
			wantOut: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsSpec(tc.input)
			if got != tc.wantOut {
				t.Errorf("expected %v, got %v", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameIsExternal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		{
			name:    "EXTERNAL reference",
			input:   "EXTERNAL/proto/api.proto",
			wantOut: true,
		},
		{
			name:    "SPEC reference",
			input:   "SPEC/x",
			wantOut: false,
		},
		{
			name:    "ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantOut: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsExternal(tc.input)
			if got != tc.wantOut {
				t.Errorf("expected %v, got %v", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameGetArtifactGenerator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		{
			name:     "Simple artifact",
			input:    "ARTIFACT/x",
			wantName: "SPEC/x",
		},
		{
			name:     "Nested artifact",
			input:    "ARTIFACT/x/y/z",
			wantName: "SPEC/x/y/z",
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
			if got != tc.wantName {
				t.Errorf("expected %q, got %q", tc.wantName, got)
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
				t.Fatal("expected non-nil PathCfs, got nil")
			}
			if got.Value != tc.wantPath {
				t.Errorf("expected %q, got %q", tc.wantPath, got.Value)
			}
		})
	}
}
