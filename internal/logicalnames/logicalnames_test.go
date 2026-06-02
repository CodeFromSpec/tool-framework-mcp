// code-from-spec: ROOT/golang/tests/utils/logical_names@jCv6y15DObkedjqXPqKnQd1aJtM
package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func TestLogicalNameToPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPath    string
		wantErr     error
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
				t.Fatal("expected non-nil result")
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
		wantPresent   bool
	}{
		{
			name:          "Extracts qualifier from ROOT reference",
			input:         "ROOT/x/y(interface)",
			wantQualifier: "interface",
			wantPresent:   true,
		},
		{
			name:        "ARTIFACT without qualifier returns absent",
			input:       "ARTIFACT/x/y",
			wantPresent: false,
		},
		{
			name:        "Returns absent when no qualifier",
			input:       "ROOT/x/y",
			wantPresent: false,
		},
		{
			name:        "Returns absent for ROOT alone",
			input:       "ROOT",
			wantPresent: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, present := logicalnames.LogicalNameGetQualifier(tc.input)
			if present != tc.wantPresent {
				t.Errorf("expected present=%v, got %v", tc.wantPresent, present)
			}
			if tc.wantPresent && got != tc.wantQualifier {
				t.Errorf("expected qualifier %q, got %q", tc.wantQualifier, got)
			}
		})
	}
}

func TestLogicalNameStripQualifier(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
	}{
		{
			name:  "Strips qualifier from ROOT reference",
			input: "ROOT/x/y(interface)",
			want:  "ROOT/x/y",
		},
		{
			name:  "ARTIFACT without qualifier — returns unchanged",
			input: "ARTIFACT/x/y",
			want:  "ARTIFACT/x/y",
		},
		{
			name:  "No qualifier — returns unchanged",
			input: "ROOT/x/y",
			want:  "ROOT/x/y",
		},
		{
			name:  "ROOT alone — returns unchanged",
			input: "ROOT",
			want:  "ROOT",
		},
		{
			name:  "Empty string — returns unchanged",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameStripQualifier(tc.input)
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
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
			name:  "ROOT alone",
			input: "ROOT",
			want:  false,
		},
		{
			name:  "ROOT with path",
			input: "ROOT/domain/config",
			want:  true,
		},
		{
			name:  "ROOT with qualifier",
			input: "ROOT/domain/config(interface)",
			want:  true,
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
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
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
			input: "ROOT/x",
			want:  false,
		},
		{
			name:  "With qualifier",
			input: "ROOT/x(y)",
			want:  true,
		},
		{
			name:  "ARTIFACT without qualifier",
			input: "ARTIFACT/x",
			want:  false,
		},
		{
			name:  "ROOT alone",
			input: "ROOT",
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
				t.Errorf("expected %v, got %v", tc.want, got)
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
			name:  "ROOT reference",
			input: "ROOT/x(y)",
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
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLogicalNameGetArtifactGenerator(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantErr   error
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
