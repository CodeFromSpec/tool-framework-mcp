// code-from-spec: ROOT/golang/tests/utils/logical_names@tizdeABgzzi2jYneADOYGweLT_E
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
	type testCase struct {
		name        string
		input       string
		wantPath    string
		wantErr     error
	}

	tests := []testCase{
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
				t.Errorf("got path %q, want %q", got.Value, tc.wantPath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameFromPath
// ---------------------------------------------------------------------------

func TestLogicalNameFromPath(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}

	tests := []testCase{
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
			cfsPath := &pathutils.PathCfs{Value: tc.input}
			got, err := logicalnames.LogicalNameFromPath(cfsPath)
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
				t.Errorf("got name %q, want %q", got, tc.wantName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameGetParent
// ---------------------------------------------------------------------------

func TestLogicalNameGetParent(t *testing.T) {
	type testCase struct {
		name       string
		input      string
		wantParent string
		wantErr    error
	}

	tests := []testCase{
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
				t.Errorf("got parent %q, want %q", got, tc.wantParent)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameGetQualifier
// ---------------------------------------------------------------------------

func TestLogicalNameGetQualifier(t *testing.T) {
	type testCase struct {
		name          string
		input         string
		wantQualifier string
		wantPresent   bool
	}

	tests := []testCase{
		// TC-16
		{
			name:          "extracts qualifier from ROOT reference",
			input:         "ROOT/x/y(interface)",
			wantQualifier: "interface",
			wantPresent:   true,
		},
		// TC-17
		{
			name:          "extracts qualifier from ARTIFACT reference",
			input:         "ARTIFACT/x/y(id)",
			wantQualifier: "id",
			wantPresent:   true,
		},
		// TC-18
		{
			name:        "returns absent when no qualifier",
			input:       "ROOT/x/y",
			wantPresent: false,
		},
		// TC-19
		{
			name:        "returns absent for ROOT alone",
			input:       "ROOT",
			wantPresent: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotQualifier, gotPresent := logicalnames.LogicalNameGetQualifier(tc.input)
			if gotPresent != tc.wantPresent {
				t.Errorf("got present=%v, want present=%v", gotPresent, tc.wantPresent)
			}
			if tc.wantPresent && gotQualifier != tc.wantQualifier {
				t.Errorf("got qualifier %q, want %q", gotQualifier, tc.wantQualifier)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameStripQualifier
// ---------------------------------------------------------------------------

func TestLogicalNameStripQualifier(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  string
	}

	tests := []testCase{
		// TC-20
		{
			name:  "strips qualifier from ROOT reference",
			input: "ROOT/x/y(interface)",
			want:  "ROOT/x/y",
		},
		// TC-21
		{
			name:  "strips qualifier from ARTIFACT reference",
			input: "ARTIFACT/x/y(id)",
			want:  "ARTIFACT/x/y",
		},
		// TC-22
		{
			name:  "no qualifier returns unchanged",
			input: "ROOT/x/y",
			want:  "ROOT/x/y",
		},
		// TC-23
		{
			name:  "ROOT alone returns unchanged",
			input: "ROOT",
			want:  "ROOT",
		},
		// TC-24
		{
			name:  "empty string returns unchanged",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameStripQualifier(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LogicalNameHasParent
// ---------------------------------------------------------------------------

func TestLogicalNameHasParent(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  bool
	}

	tests := []testCase{
		// TC-25
		{
			name:  "ROOT alone",
			input: "ROOT",
			want:  false,
		},
		// TC-26
		{
			name:  "ROOT with path",
			input: "ROOT/domain/config",
			want:  true,
		},
		// TC-27
		{
			name:  "ROOT with qualifier",
			input: "ROOT/domain/config(interface)",
			want:  true,
		},
		// TC-28
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x(y)",
			want:  false,
		},
		// TC-29
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
	type testCase struct {
		name  string
		input string
		want  bool
	}

	tests := []testCase{
		// TC-30
		{
			name:  "without qualifier",
			input: "ROOT/x",
			want:  false,
		},
		// TC-31
		{
			name:  "with qualifier",
			input: "ROOT/x(y)",
			want:  true,
		},
		// TC-32
		{
			name:  "ARTIFACT with qualifier",
			input: "ARTIFACT/x(y)",
			want:  true,
		},
		// TC-33
		{
			name:  "ROOT alone",
			input: "ROOT",
			want:  false,
		},
		// TC-34
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
	type testCase struct {
		name  string
		input string
		want  bool
	}

	tests := []testCase{
		// TC-35
		{
			name:  "ARTIFACT reference",
			input: "ARTIFACT/x(y)",
			want:  true,
		},
		// TC-36
		{
			name:  "ROOT reference",
			input: "ROOT/x(y)",
			want:  false,
		},
		// TC-37
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
	type testCase struct {
		name      string
		input     string
		wantName  string
		wantErr   error
	}

	tests := []testCase{
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
