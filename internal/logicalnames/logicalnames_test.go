// code-from-spec: ROOT/golang/tests/utils/logical_names@tIIp-kUAPDumljUoPjqBW6-hm38

package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ---------------------------------------------------------------------------
// LogicalNameToPath
// ---------------------------------------------------------------------------

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
			name:     "strips qualifier before resolving",
			input:    "ROOT/x/y(interface)",
			wantPath: "code-from-spec/x/y/_node.md",
		},
		{
			name:    "rejects ARTIFACT reference",
			input:   "ARTIFACT/x(y)",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
		{
			name:    "rejects unrecognized prefix",
			input:   "UNKNOWN/something",
			wantErr: logicalnames.ErrUnsupportedReference,
		},
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
			if got.Value != tc.wantPath {
				t.Errorf("got %q, want %q", got.Value, tc.wantPath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameFromPath
// ---------------------------------------------------------------------------

func TestLogicalNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		{
			name:     "root node",
			input:    "code-from-spec/_node.md",
			wantName: "ROOT",
		},
		{
			name:     "nested node",
			input:    "code-from-spec/x/y/_node.md",
			wantName: "ROOT/x/y",
		},
		{
			name:    "rejects non-node path",
			input:   "internal/config/config.go",
			wantErr: logicalnames.ErrInvalidPath,
		},
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
				t.Errorf("got %q, want %q", got, tc.wantName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameGetParent
// ---------------------------------------------------------------------------

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
			name:       "strips qualifier before computing parent",
			input:      "ROOT/domain/config(interface)",
			wantParent: "ROOT/domain",
		},
		{
			name:    "ROOT has no parent",
			input:   "ROOT",
			wantErr: logicalnames.ErrNoParent,
		},
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
				t.Errorf("got %q, want %q", got, tc.wantParent)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameGetQualifier
// ---------------------------------------------------------------------------

func TestLogicalNameGetQualifier(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantQualifier string
		wantOk        bool
	}{
		{
			name:          "extracts qualifier from ROOT reference",
			input:         "ROOT/x/y(interface)",
			wantQualifier: "interface",
			wantOk:        true,
		},
		{
			name:          "extracts qualifier from ARTIFACT reference",
			input:         "ARTIFACT/x/y(id)",
			wantQualifier: "id",
			wantOk:        true,
		},
		{
			name:   "returns absent when no qualifier",
			input:  "ROOT/x/y",
			wantOk: false,
		},
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
				t.Fatalf("ok: got %v, want %v", ok, tc.wantOk)
			}
			if tc.wantOk && got != tc.wantQualifier {
				t.Errorf("qualifier: got %q, want %q", got, tc.wantQualifier)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameStripQualifier
// ---------------------------------------------------------------------------

func TestLogicalNameStripQualifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOut  string
	}{
		{
			name:    "strips qualifier from ROOT reference",
			input:   "ROOT/x/y(interface)",
			wantOut: "ROOT/x/y",
		},
		{
			name:    "strips qualifier from ARTIFACT reference",
			input:   "ARTIFACT/x/y(id)",
			wantOut: "ARTIFACT/x/y",
		},
		{
			name:    "no qualifier — returns unchanged",
			input:   "ROOT/x/y",
			wantOut: "ROOT/x/y",
		},
		{
			name:    "ROOT alone — returns unchanged",
			input:   "ROOT",
			wantOut: "ROOT",
		},
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
				t.Errorf("got %q, want %q", got, tc.wantOut)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameHasParent
// ---------------------------------------------------------------------------

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
			input: "ARTIFACT/x(y)",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameHasQualifier
// ---------------------------------------------------------------------------

func TestLogicalNameHasQualifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "without qualifier",
			input: "ROOT/x",
			want:  false,
		},
		{
			name:  "with qualifier",
			input: "ROOT/x(y)",
			want:  true,
		},
		{
			name:  "ARTIFACT with qualifier",
			input: "ARTIFACT/x(y)",
			want:  true,
		},
		{
			name:  "ROOT alone",
			input: "ROOT",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasQualifier(tc.input)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameIsArtifact
// ---------------------------------------------------------------------------

func TestLogicalNameIsArtifact(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x(y)",
			want:  true,
		},
		{
			name:  "ROOT reference",
			input: "ROOT/x(y)",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsArtifact(tc.input)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameGetArtifactGenerator
// ---------------------------------------------------------------------------

func TestLogicalNameGetArtifactGenerator(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantGen   string
		wantErr   error
	}{
		{
			name:    "simple artifact",
			input:   "ARTIFACT/x(y)",
			wantGen: "ROOT/x",
		},
		{
			name:    "nested artifact",
			input:   "ARTIFACT/x/y/z(id)",
			wantGen: "ROOT/x/y/z",
		},
		{
			name:    "rejects ROOT reference",
			input:   "ROOT/x(y)",
			wantErr: logicalnames.ErrNotAnArtifactReference,
		},
		{
			name:    "artifact reference without qualifier",
			input:   "ARTIFACT/x",
			wantGen: "ROOT/x",
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
			if got != tc.wantGen {
				t.Errorf("got %q, want %q", got, tc.wantGen)
			}
		})
	}
}
