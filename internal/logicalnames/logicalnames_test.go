// code-from-spec: ROOT/golang/tests/utils/logical_names@47TIH-NvzCl_XAxAv7oQdlKpWpA
package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// --- LogicalNameToPath ---

func TestLogicalNameToPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPath    string
		wantErr     error
	}{
		// TC-01
		{
			name:     "ROOT alone",
			input:    "ROOT",
			wantPath: "code-from-spec/_node.md",
		},
		// TC-02
		{
			name:     "ROOT with path",
			input:    "ROOT/payments/processor",
			wantPath: "code-from-spec/payments/processor/_node.md",
		},
		// TC-03
		{
			name:     "strips qualifier before resolving",
			input:    "ROOT/x/y(interface)",
			wantPath: "code-from-spec/x/y/_node.md",
		},
		// TC-04
		{
			name:    "rejects ARTIFACT reference",
			input:   "ARTIFACT/x(y)",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		// TC-05
		{
			name:    "rejects unrecognized prefix",
			input:   "UNKNOWN/something",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		// TC-06
		{
			name:    "rejects empty string",
			input:   "",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameToPath(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("LogicalNameToPath(%q) error = %v, want %v", tc.input, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LogicalNameToPath(%q) unexpected error: %v", tc.input, err)
			}
			if got.Value != tc.wantPath {
				t.Errorf("LogicalNameToPath(%q) = %q, want %q", tc.input, got.Value, tc.wantPath)
			}
		})
	}
}

// --- LogicalNameFromPath ---

func TestLogicalNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		// TC-07
		{
			name:     "root node",
			input:    "code-from-spec/_node.md",
			wantName: "ROOT",
		},
		// TC-08
		{
			name:     "nested node",
			input:    "code-from-spec/x/y/_node.md",
			wantName: "ROOT/x/y",
		},
		// TC-09
		{
			name:    "rejects non-node path",
			input:   "internal/config/config.go",
			wantErr: logicalnames.ErrInvalidPath,
		},
		// TC-10
		{
			name:    "rejects path without _node.md",
			input:   "code-from-spec/x/y/output.md",
			wantErr: logicalnames.ErrInvalidPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: tc.input})
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("LogicalNameFromPath(%q) error = %v, want %v", tc.input, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LogicalNameFromPath(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.wantName {
				t.Errorf("LogicalNameFromPath(%q) = %q, want %q", tc.input, got, tc.wantName)
			}
		})
	}
}

// --- LogicalNameGetParent ---

func TestLogicalNameGetParent(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantParent string
		wantErr    error
	}{
		// TC-11
		{
			name:       "ROOT/x parent is ROOT",
			input:      "ROOT/domain",
			wantParent: "ROOT",
		},
		// TC-12
		{
			name:       "ROOT/x/y parent is ROOT/x",
			input:      "ROOT/domain/config",
			wantParent: "ROOT/domain",
		},
		// TC-13
		{
			name:       "strips qualifier before computing parent",
			input:      "ROOT/domain/config(interface)",
			wantParent: "ROOT/domain",
		},
		// TC-14
		{
			name:    "ROOT has no parent",
			input:   "ROOT",
			wantErr: logicalnames.ErrNoParent,
		},
		// TC-15
		{
			name:    "rejects ARTIFACT reference",
			input:   "ARTIFACT/x(y)",
			wantErr: logicalnames.ErrNotARootReference,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetParent(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("LogicalNameGetParent(%q) error = %v, want %v", tc.input, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LogicalNameGetParent(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.wantParent {
				t.Errorf("LogicalNameGetParent(%q) = %q, want %q", tc.input, got, tc.wantParent)
			}
		})
	}
}

// --- LogicalNameGetQualifier ---

func TestLogicalNameGetQualifier(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantQualifier string
		wantOk        bool
	}{
		// TC-16
		{
			name:          "extracts qualifier from ROOT reference",
			input:         "ROOT/x/y(interface)",
			wantQualifier: "interface",
			wantOk:        true,
		},
		// TC-17
		{
			name:          "extracts qualifier from ARTIFACT reference",
			input:         "ARTIFACT/x/y(id)",
			wantQualifier: "id",
			wantOk:        true,
		},
		// TC-18
		{
			name:   "returns absent when no qualifier",
			input:  "ROOT/x/y",
			wantOk: false,
		},
		// TC-19
		{
			name:   "returns absent for ROOT alone",
			input:  "ROOT",
			wantOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := logicalnames.LogicalNameGetQualifier(tc.input)
			if ok != tc.wantOk {
				t.Errorf("LogicalNameGetQualifier(%q) ok = %v, want %v", tc.input, ok, tc.wantOk)
			}
			if ok && got != tc.wantQualifier {
				t.Errorf("LogicalNameGetQualifier(%q) qualifier = %q, want %q", tc.input, got, tc.wantQualifier)
			}
		})
	}
}

// --- LogicalNameStripQualifier ---

func TestLogicalNameStripQualifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOut  string
	}{
		// TC-20
		{
			name:    "strips qualifier from ROOT reference",
			input:   "ROOT/x/y(interface)",
			wantOut: "ROOT/x/y",
		},
		// TC-21
		{
			name:    "strips qualifier from ARTIFACT reference",
			input:   "ARTIFACT/x/y(id)",
			wantOut: "ARTIFACT/x/y",
		},
		// TC-22
		{
			name:    "no qualifier — returns unchanged",
			input:   "ROOT/x/y",
			wantOut: "ROOT/x/y",
		},
		// TC-23
		{
			name:    "ROOT alone — returns unchanged",
			input:   "ROOT",
			wantOut: "ROOT",
		},
		// TC-24
		{
			name:    "empty string — returns unchanged",
			input:   "",
			wantOut: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameStripQualifier(tc.input)
			if got != tc.wantOut {
				t.Errorf("LogicalNameStripQualifier(%q) = %q, want %q", tc.input, got, tc.wantOut)
			}
		})
	}
}

// --- LogicalNameHasParent ---

func TestLogicalNameHasParent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		// TC-25
		{
			name:    "ROOT alone",
			input:   "ROOT",
			wantOut: false,
		},
		// TC-26
		{
			name:    "ROOT with path",
			input:   "ROOT/domain/config",
			wantOut: true,
		},
		// TC-27
		{
			name:    "ROOT with qualifier",
			input:   "ROOT/domain/config(interface)",
			wantOut: true,
		},
		// TC-28
		{
			name:    "ARTIFACT reference",
			input:   "ARTIFACT/x(y)",
			wantOut: false,
		},
		// TC-29
		{
			name:    "empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.wantOut {
				t.Errorf("LogicalNameHasParent(%q) = %v, want %v", tc.input, got, tc.wantOut)
			}
		})
	}
}

// --- LogicalNameHasQualifier ---

func TestLogicalNameHasQualifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		// TC-30
		{
			name:    "without qualifier",
			input:   "ROOT/x",
			wantOut: false,
		},
		// TC-31
		{
			name:    "with qualifier",
			input:   "ROOT/x(y)",
			wantOut: true,
		},
		// TC-32
		{
			name:    "ARTIFACT with qualifier",
			input:   "ARTIFACT/x(y)",
			wantOut: true,
		},
		// TC-33
		{
			name:    "ROOT alone",
			input:   "ROOT",
			wantOut: false,
		},
		// TC-34
		{
			name:    "empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasQualifier(tc.input)
			if got != tc.wantOut {
				t.Errorf("LogicalNameHasQualifier(%q) = %v, want %v", tc.input, got, tc.wantOut)
			}
		})
	}
}

// --- LogicalNameIsArtifact ---

func TestLogicalNameIsArtifact(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut bool
	}{
		// TC-35
		{
			name:    "ARTIFACT reference",
			input:   "ARTIFACT/x(y)",
			wantOut: true,
		},
		// TC-36
		{
			name:    "ROOT reference",
			input:   "ROOT/x(y)",
			wantOut: false,
		},
		// TC-37
		{
			name:    "empty string",
			input:   "",
			wantOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsArtifact(tc.input)
			if got != tc.wantOut {
				t.Errorf("LogicalNameIsArtifact(%q) = %v, want %v", tc.input, got, tc.wantOut)
			}
		})
	}
}

// --- LogicalNameGetArtifactGenerator ---

func TestLogicalNameGetArtifactGenerator(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantErr   error
	}{
		// TC-38
		{
			name:     "simple artifact",
			input:    "ARTIFACT/x(y)",
			wantName: "ROOT/x",
		},
		// TC-39
		{
			name:     "nested artifact",
			input:    "ARTIFACT/x/y/z(id)",
			wantName: "ROOT/x/y/z",
		},
		// TC-40
		{
			name:    "rejects ROOT reference",
			input:   "ROOT/x(y)",
			wantErr: logicalnames.ErrNotAnArtifactReference,
		},
		// TC-41
		{
			name:     "artifact reference without qualifier",
			input:    "ARTIFACT/x",
			wantName: "ROOT/x",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetArtifactGenerator(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("LogicalNameGetArtifactGenerator(%q) error = %v, want %v", tc.input, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LogicalNameGetArtifactGenerator(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.wantName {
				t.Errorf("LogicalNameGetArtifactGenerator(%q) = %q, want %q", tc.input, got, tc.wantName)
			}
		})
	}
}
