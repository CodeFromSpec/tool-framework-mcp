// code-from-spec: ROOT/golang/tests/utils/logical_names@_usSKmUSLkcD_dKb8IDAHnWITaI
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
			name:     "ROOT alone",
			input:    "ROOT",
			wantPath: "code-from-spec/_node.md",
		},
		{
			name:     "ROOT with path",
			input:    "ROOT/payments/processor",
			wantPath: "code-from-spec/payments/processor/_node.md",
		},
		{
			name:     "Strips qualifier before resolving",
			input:    "ROOT/x/y(interface)",
			wantPath: "code-from-spec/x/y/_node.md",
		},
		{
			name:    "Rejects ARTIFACT reference",
			input:   "ARTIFACT/x",
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
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Value != tc.wantPath {
				t.Fatalf("expected %q, got %q", tc.wantPath, got.Value)
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
			wantName: "ROOT",
		},
		{
			name:     "Nested node",
			input:    "code-from-spec/x/y/_node.md",
			wantName: "ROOT/x/y",
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
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantName {
				t.Fatalf("expected %q, got %q", tc.wantName, got)
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
			name:       "ROOT/x parent is ROOT",
			input:      "ROOT/domain",
			wantParent: "ROOT",
		},
		{
			name:       "ROOT/x/y parent is ROOT/x",
			input:      "ROOT/domain/config",
			wantParent: "ROOT/domain",
		},
		{
			name:       "Strips qualifier before computing parent",
			input:      "ROOT/domain/config(interface)",
			wantParent: "ROOT/domain",
		},
		{
			name:    "ROOT has no parent",
			input:   "ROOT",
			wantErr: logicalnames.ErrNoParent,
		},
		{
			name:    "Rejects ARTIFACT reference",
			input:   "ARTIFACT/x",
			wantErr: logicalnames.ErrNotARootReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetParent(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantParent {
				t.Fatalf("expected %q, got %q", tc.wantParent, got)
			}
		})
	}
}

func TestLogicalNameGetQualifier(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantQ     string
		wantFound bool
	}{
		{
			name:      "Extracts qualifier from ROOT reference",
			input:     "ROOT/x/y(interface)",
			wantQ:     "interface",
			wantFound: true,
		},
		{
			name:      "ARTIFACT without qualifier returns absent",
			input:     "ARTIFACT/x/y",
			wantFound: false,
		},
		{
			name:      "Returns absent when no qualifier",
			input:     "ROOT/x/y",
			wantFound: false,
		},
		{
			name:      "Returns absent for ROOT alone",
			input:     "ROOT",
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, found := logicalnames.LogicalNameGetQualifier(tc.input)
			if found != tc.wantFound {
				t.Fatalf("expected found=%v, got found=%v", tc.wantFound, found)
			}
			if found && got != tc.wantQ {
				t.Fatalf("expected qualifier %q, got %q", tc.wantQ, got)
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
			name:    "Strips qualifier from ROOT reference",
			input:   "ROOT/x/y(interface)",
			wantOut: "ROOT/x/y",
		},
		{
			name:    "ARTIFACT without qualifier returns unchanged",
			input:   "ARTIFACT/x/y",
			wantOut: "ARTIFACT/x/y",
		},
		{
			name:    "No qualifier returns unchanged",
			input:   "ROOT/x/y",
			wantOut: "ROOT/x/y",
		},
		{
			name:    "ROOT alone returns unchanged",
			input:   "ROOT",
			wantOut: "ROOT",
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
				t.Fatalf("expected %q, got %q", tc.wantOut, got)
			}
		})
	}
}

func TestLogicalNameHasParent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBool bool
	}{
		{
			name:     "ROOT alone",
			input:    "ROOT",
			wantBool: false,
		},
		{
			name:     "ROOT with path",
			input:    "ROOT/domain/config",
			wantBool: true,
		},
		{
			name:     "ROOT with qualifier",
			input:    "ROOT/domain/config(interface)",
			wantBool: true,
		},
		{
			name:     "ARTIFACT reference",
			input:    "ARTIFACT/x",
			wantBool: false,
		},
		{
			name:     "Empty string",
			input:    "",
			wantBool: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.wantBool {
				t.Fatalf("expected %v, got %v", tc.wantBool, got)
			}
		})
	}
}

func TestLogicalNameHasQualifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBool bool
	}{
		{
			name:     "Without qualifier",
			input:    "ROOT/x",
			wantBool: false,
		},
		{
			name:     "With qualifier",
			input:    "ROOT/x(y)",
			wantBool: true,
		},
		{
			name:     "ARTIFACT without qualifier",
			input:    "ARTIFACT/x",
			wantBool: false,
		},
		{
			name:     "ROOT alone",
			input:    "ROOT",
			wantBool: false,
		},
		{
			name:     "Empty string",
			input:    "",
			wantBool: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasQualifier(tc.input)
			if got != tc.wantBool {
				t.Fatalf("expected %v, got %v", tc.wantBool, got)
			}
		})
	}
}

func TestLogicalNameIsArtifact(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBool bool
	}{
		{
			name:     "ARTIFACT reference",
			input:    "ARTIFACT/x",
			wantBool: true,
		},
		{
			name:     "ROOT reference",
			input:    "ROOT/x(y)",
			wantBool: false,
		},
		{
			name:     "Empty string",
			input:    "",
			wantBool: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsArtifact(tc.input)
			if got != tc.wantBool {
				t.Fatalf("expected %v, got %v", tc.wantBool, got)
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
			wantName: "ROOT/x",
		},
		{
			name:     "Nested artifact",
			input:    "ARTIFACT/x/y/z",
			wantName: "ROOT/x/y/z",
		},
		{
			name:    "Rejects ROOT reference",
			input:   "ROOT/x(y)",
			wantErr: logicalnames.ErrNotAnArtifactReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetArtifactGenerator(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantName {
				t.Fatalf("expected %q, got %q", tc.wantName, got)
			}
		})
	}
}
