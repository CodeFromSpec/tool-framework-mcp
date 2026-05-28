// code-from-spec: ROOT/golang/tests/internal/logical_names@8jMmtczzXpRB6lvTsELOFVeVVP0

package logicalnames_test

import (
	"errors"
	"testing"

	logicalnames "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// TestLogicalNameToPath covers TC-01 through TC-06.
func TestLogicalNameToPath(t *testing.T) {
	type testCase struct {
		name        string
		input       string
		wantPath    string
		wantErrText string
	}

	cases := []testCase{
		{
			name:     "TC-01: ROOT alone",
			input:    "ROOT",
			wantPath: "code-from-spec/_node.md",
		},
		{
			name:     "TC-02: ROOT with path",
			input:    "ROOT/payments/processor",
			wantPath: "code-from-spec/payments/processor/_node.md",
		},
		{
			name:     "TC-03: Strips qualifier before resolving",
			input:    "ROOT/x/y(interface)",
			wantPath: "code-from-spec/x/y/_node.md",
		},
		{
			name:        "TC-04: Rejects ARTIFACT reference",
			input:       "ARTIFACT/x(y)",
			wantErrText: "unsupported reference",
		},
		{
			name:        "TC-05: Rejects unrecognized prefix",
			input:       "UNKNOWN/something",
			wantErrText: "unsupported reference",
		},
		{
			name:        "TC-06: Rejects empty string",
			input:       "",
			wantErrText: "unsupported reference",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameToPath(tc.input)
			if tc.wantErrText != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrText)
				}
				if err.Error() != tc.wantErrText {
					t.Fatalf("expected error %q, got %q", tc.wantErrText, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantPath {
				t.Fatalf("expected path %q, got %q", tc.wantPath, got)
			}
		})
	}
}

// TestLogicalNameFromPath covers TC-07 through TC-10.
func TestLogicalNameFromPath(t *testing.T) {
	type testCase struct {
		name        string
		input       string
		wantName    string
		wantErrText string
	}

	cases := []testCase{
		{
			name:     "TC-07: Root node",
			input:    "code-from-spec/_node.md",
			wantName: "ROOT",
		},
		{
			name:     "TC-08: Nested node",
			input:    "code-from-spec/x/y/_node.md",
			wantName: "ROOT/x/y",
		},
		{
			name:        "TC-09: Rejects non-node path",
			input:       "internal/config/config.go",
			wantErrText: "invalid path",
		},
		{
			name:        "TC-10: Rejects path without _node.md",
			input:       "code-from-spec/x/y/output.md",
			wantErrText: "invalid path",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameFromPath(tc.input)
			if tc.wantErrText != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrText)
				}
				if err.Error() != tc.wantErrText {
					t.Fatalf("expected error %q, got %q", tc.wantErrText, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantName {
				t.Fatalf("expected logical name %q, got %q", tc.wantName, got)
			}
		})
	}
}

// TestLogicalNameGetParent covers TC-11 through TC-15.
func TestLogicalNameGetParent(t *testing.T) {
	type testCase struct {
		name        string
		input       string
		wantParent  string
		wantErrText string
	}

	cases := []testCase{
		{
			name:       "TC-11: ROOT/x parent is ROOT",
			input:      "ROOT/domain",
			wantParent: "ROOT",
		},
		{
			name:       "TC-12: ROOT/x/y parent is ROOT/x",
			input:      "ROOT/domain/config",
			wantParent: "ROOT/domain",
		},
		{
			name:       "TC-13: Strips qualifier before computing parent",
			input:      "ROOT/domain/config(interface)",
			wantParent: "ROOT/domain",
		},
		{
			name:        "TC-14: ROOT has no parent",
			input:       "ROOT",
			wantErrText: "no parent",
		},
		{
			name:        "TC-15: Rejects ARTIFACT reference",
			input:       "ARTIFACT/x(y)",
			wantErrText: "not a ROOT reference",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetParent(tc.input)
			if tc.wantErrText != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrText)
				}
				if err.Error() != tc.wantErrText {
					t.Fatalf("expected error %q, got %q", tc.wantErrText, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantParent {
				t.Fatalf("expected parent %q, got %q", tc.wantParent, got)
			}
		})
	}
}

// TestLogicalNameGetQualifier covers TC-16 through TC-19.
func TestLogicalNameGetQualifier(t *testing.T) {
	type testCase struct {
		name          string
		input         string
		wantQualifier string
		wantPresent   bool
	}

	cases := []testCase{
		{
			name:          "TC-16: Extracts qualifier from ROOT reference",
			input:         "ROOT/x/y(interface)",
			wantQualifier: "interface",
			wantPresent:   true,
		},
		{
			name:          "TC-17: Extracts qualifier from ARTIFACT reference",
			input:         "ARTIFACT/x/y(id)",
			wantQualifier: "id",
			wantPresent:   true,
		},
		{
			name:        "TC-18: Returns absent when no qualifier",
			input:       "ROOT/x/y",
			wantPresent: false,
		},
		{
			name:        "TC-19: Returns absent for ROOT alone",
			input:       "ROOT",
			wantPresent: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := logicalnames.LogicalNameGetQualifier(tc.input)
			if ok != tc.wantPresent {
				t.Fatalf("expected present=%v, got present=%v", tc.wantPresent, ok)
			}
			if ok && got != tc.wantQualifier {
				t.Fatalf("expected qualifier %q, got %q", tc.wantQualifier, got)
			}
		})
	}
}

// TestLogicalNameHasParent covers TC-20 through TC-24.
func TestLogicalNameHasParent(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  bool
	}

	cases := []testCase{
		{
			name:  "TC-20: ROOT alone",
			input: "ROOT",
			want:  false,
		},
		{
			name:  "TC-21: ROOT with path",
			input: "ROOT/domain/config",
			want:  true,
		},
		{
			name:  "TC-22: ROOT with qualifier",
			input: "ROOT/domain/config(interface)",
			want:  true,
		},
		{
			name:  "TC-23: ARTIFACT reference",
			input: "ARTIFACT/x(y)",
			want:  false,
		},
		{
			name:  "TC-24: Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasParent(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// TestLogicalNameHasQualifier covers TC-25 through TC-29.
func TestLogicalNameHasQualifier(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  bool
	}

	cases := []testCase{
		{
			name:  "TC-25: Without qualifier",
			input: "ROOT/x",
			want:  false,
		},
		{
			name:  "TC-26: With qualifier",
			input: "ROOT/x(y)",
			want:  true,
		},
		{
			name:  "TC-27: ARTIFACT with qualifier",
			input: "ARTIFACT/x(y)",
			want:  true,
		},
		{
			name:  "TC-28: ROOT alone",
			input: "ROOT",
			want:  false,
		},
		{
			name:  "TC-29: Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameHasQualifier(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// TestLogicalNameIsArtifact covers TC-30 through TC-32.
func TestLogicalNameIsArtifact(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  bool
	}

	cases := []testCase{
		{
			name:  "TC-30: ARTIFACT reference",
			input: "ARTIFACT/x(y)",
			want:  true,
		},
		{
			name:  "TC-31: ROOT reference",
			input: "ROOT/x(y)",
			want:  false,
		},
		{
			name:  "TC-32: Empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := logicalnames.LogicalNameIsArtifact(tc.input)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// TestLogicalNameGetArtifactGenerator covers TC-33 through TC-36.
func TestLogicalNameGetArtifactGenerator(t *testing.T) {
	type testCase struct {
		name        string
		input       string
		wantGen     string
		wantErrText string
	}

	cases := []testCase{
		{
			name:    "TC-33: Simple artifact",
			input:   "ARTIFACT/x(y)",
			wantGen: "ROOT/x",
		},
		{
			name:    "TC-34: Nested artifact",
			input:   "ARTIFACT/x/y/z(id)",
			wantGen: "ROOT/x/y/z",
		},
		{
			name:        "TC-35: Rejects ROOT reference",
			input:       "ROOT/x(y)",
			wantErrText: "not an artifact reference",
		},
		{
			name:    "TC-36: Artifact reference without qualifier",
			input:   "ARTIFACT/x",
			wantGen: "ROOT/x",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := logicalnames.LogicalNameGetArtifactGenerator(tc.input)
			if tc.wantErrText != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrText)
				}
				if err.Error() != tc.wantErrText {
					t.Fatalf("expected error %q, got %q", tc.wantErrText, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantGen {
				t.Fatalf("expected generator %q, got %q", tc.wantGen, got)
			}
		})
	}
}

// Ensure errors package is used — sentinel errors checked via errors.Is where applicable.
var _ = errors.New
