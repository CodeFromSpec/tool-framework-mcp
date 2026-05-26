// code-from-spec: ROOT/golang/internal/logical_names/tests@QMiSLlIcVJHpJx30FVzMTXxIyhA

// Package logicalnames — test file for all pure functions exported by this package.
//
// These are pure-function tests: no filesystem or temp directories are needed.
// Each test calls a function with a string input and asserts the output against
// the expected values declared in the spec.
//
// Conventions:
//   - Table-driven tests using a local testCase struct per function group.
//   - All helper types/functions are prefixed with "test" to avoid collisions
//     with unexported names in the package under test.
//   - No external test frameworks — only the standard "testing" package.
package logicalnames

import "testing"

// ─── PathFromLogicalName ──────────────────────────────────────────────────────

func TestPathFromLogicalName(t *testing.T) {
	// testCase holds one row of the table for PathFromLogicalName.
	type testCase struct {
		name   string // human-readable description
		input  string // logical name passed to PathFromLogicalName
		wantPath string // expected returned file path
		wantOK   bool   // expected second return value
	}

	cases := []testCase{
		{
			// The root node itself.
			name:     "ROOT",
			input:    "ROOT",
			wantPath: "code-from-spec/_node.md",
			wantOK:   true,
		},
		{
			// Multi-segment ROOT path with no qualifier.
			name:     "ROOT with path",
			input:    "ROOT/payments/processor",
			wantPath: "code-from-spec/payments/processor/_node.md",
			wantOK:   true,
		},
		{
			// Multi-segment ROOT path with a qualifier — qualifier must be stripped.
			name:     "ROOT with qualifier",
			input:    "ROOT/payments/processor(interface)",
			wantPath: "code-from-spec/payments/processor/_node.md",
			wantOK:   true,
		},
		{
			// Single-segment ROOT path with a qualifier — exercises the strip path.
			name:     "ROOT with qualifier strips qualifier from path",
			input:    "ROOT/x(y)",
			wantPath: "code-from-spec/x/_node.md",
			wantOK:   true,
		},
		{
			// ARTIFACT/ references are not handled by this function.
			name:     "ARTIFACT reference returns false",
			input:    "ARTIFACT/x(y)",
			wantPath: "",
			wantOK:   false,
		},
		{
			// Completely unknown prefix.
			name:     "Unrecognized prefix",
			input:    "UNKNOWN/something",
			wantPath: "",
			wantOK:   false,
		},
		{
			// Empty string is never valid.
			name:     "Empty string",
			input:    "",
			wantPath: "",
			wantOK:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotOK := PathFromLogicalName(tc.input)
			if gotPath != tc.wantPath || gotOK != tc.wantOK {
				t.Errorf(
					"PathFromLogicalName(%q) = (%q, %v), want (%q, %v)",
					tc.input, gotPath, gotOK, tc.wantPath, tc.wantOK,
				)
			}
		})
	}
}

// ─── HasParent ────────────────────────────────────────────────────────────────

func TestHasParent(t *testing.T) {
	type testCase struct {
		name          string
		input         string
		wantHasParent bool
		wantOK        bool
	}

	cases := []testCase{
		{
			// ROOT is the top of the tree — no parent.
			name:          "ROOT",
			input:         "ROOT",
			wantHasParent: false,
			wantOK:        true,
		},
		{
			// A deeper node always has a parent.
			name:          "ROOT with path",
			input:         "ROOT/domain/config",
			wantHasParent: true,
			wantOK:        true,
		},
		{
			// A qualified node also has a parent.
			name:          "ROOT with qualifier",
			input:         "ROOT/domain/config(interface)",
			wantHasParent: true,
			wantOK:        true,
		},
		{
			// ARTIFACT/ names are not valid for HasParent.
			name:          "ARTIFACT returns false false",
			input:         "ARTIFACT/x(y)",
			wantHasParent: false,
			wantOK:        false,
		},
		{
			// Empty string is not a valid logical name.
			name:          "Empty string",
			input:         "",
			wantHasParent: false,
			wantOK:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotHasParent, gotOK := HasParent(tc.input)
			if gotHasParent != tc.wantHasParent || gotOK != tc.wantOK {
				t.Errorf(
					"HasParent(%q) = (%v, %v), want (%v, %v)",
					tc.input, gotHasParent, gotOK, tc.wantHasParent, tc.wantOK,
				)
			}
		})
	}
}

// ─── ParentLogicalName ────────────────────────────────────────────────────────

func TestParentLogicalName(t *testing.T) {
	type testCase struct {
		name       string
		input      string
		wantParent string
		wantOK     bool
	}

	cases := []testCase{
		{
			// Single segment under ROOT — parent is ROOT.
			name:       "ROOT/x — parent is ROOT",
			input:      "ROOT/domain",
			wantParent: "ROOT",
			wantOK:     true,
		},
		{
			// Two segments — parent loses the last segment.
			name:       "ROOT/x/y — parent is ROOT/x",
			input:      "ROOT/domain/config",
			wantParent: "ROOT/domain",
			wantOK:     true,
		},
		{
			// Qualified name — qualifier is stripped, then last segment removed.
			name:       "ROOT/x/y(z) — parent is ROOT/x",
			input:      "ROOT/domain/config(interface)",
			wantParent: "ROOT/domain",
			wantOK:     true,
		},
		{
			// ROOT itself has no parent.
			name:       "ROOT has no parent",
			input:      "ROOT",
			wantParent: "",
			wantOK:     false,
		},
		{
			// Empty string is invalid.
			name:       "Empty string invalid",
			input:      "",
			wantParent: "",
			wantOK:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotParent, gotOK := ParentLogicalName(tc.input)
			if gotParent != tc.wantParent || gotOK != tc.wantOK {
				t.Errorf(
					"ParentLogicalName(%q) = (%q, %v), want (%q, %v)",
					tc.input, gotParent, gotOK, tc.wantParent, tc.wantOK,
				)
			}
		})
	}
}

// ─── HasQualifier ─────────────────────────────────────────────────────────────

func TestHasQualifier(t *testing.T) {
	type testCase struct {
		name             string
		input            string
		wantHasQualifier bool
		wantOK           bool
	}

	cases := []testCase{
		{
			// A ROOT/ name without any parenthetical.
			name:             "ROOT without qualifier",
			input:            "ROOT/x",
			wantHasQualifier: false,
			wantOK:           true,
		},
		{
			// A ROOT/ name with a qualifier.
			name:             "ROOT with qualifier",
			input:            "ROOT/x(y)",
			wantHasQualifier: true,
			wantOK:           true,
		},
		{
			// An ARTIFACT/ name with a qualifier (artifact id).
			name:             "ARTIFACT with qualifier",
			input:            "ARTIFACT/x(y)",
			wantHasQualifier: true,
			wantOK:           true,
		},
		{
			// ROOT alone (no path, no qualifier).
			name:             "ROOT alone",
			input:            "ROOT",
			wantHasQualifier: false,
			wantOK:           true,
		},
		{
			// Empty string is not a recognized logical name.
			name:             "Empty string",
			input:            "",
			wantHasQualifier: false,
			wantOK:           false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotHasQualifier, gotOK := HasQualifier(tc.input)
			if gotHasQualifier != tc.wantHasQualifier || gotOK != tc.wantOK {
				t.Errorf(
					"HasQualifier(%q) = (%v, %v), want (%v, %v)",
					tc.input, gotHasQualifier, gotOK, tc.wantHasQualifier, tc.wantOK,
				)
			}
		})
	}
}

// ─── QualifierName ────────────────────────────────────────────────────────────

func TestQualifierName(t *testing.T) {
	type testCase struct {
		name          string
		input         string
		wantQualifier string
		wantOK        bool
	}

	cases := []testCase{
		{
			// Single-segment ROOT with a qualifier.
			name:          "ROOT with qualifier",
			input:         "ROOT/x(y)",
			wantQualifier: "y",
			wantOK:        true,
		},
		{
			// Multi-segment ROOT with a qualifier.
			name:          "ROOT with nested path and qualifier",
			input:         "ROOT/x/y(interface)",
			wantQualifier: "interface",
			wantOK:        true,
		},
		{
			// ARTIFACT/ name — qualifier is the artifact id.
			name:          "ARTIFACT with qualifier",
			input:         "ARTIFACT/x(y)",
			wantQualifier: "y",
			wantOK:        true,
		},
		{
			// No parenthetical at all.
			name:          "ROOT without qualifier",
			input:         "ROOT/x",
			wantQualifier: "",
			wantOK:        false,
		},
		{
			// ROOT alone — no qualifier possible.
			name:          "ROOT alone",
			input:         "ROOT",
			wantQualifier: "",
			wantOK:        false,
		},
		{
			// Empty string.
			name:          "Empty string",
			input:         "",
			wantQualifier: "",
			wantOK:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotQualifier, gotOK := QualifierName(tc.input)
			if gotQualifier != tc.wantQualifier || gotOK != tc.wantOK {
				t.Errorf(
					"QualifierName(%q) = (%q, %v), want (%q, %v)",
					tc.input, gotQualifier, gotOK, tc.wantQualifier, tc.wantOK,
				)
			}
		})
	}
}

// ─── IsArtifactRef ────────────────────────────────────────────────────────────

func TestIsArtifactRef(t *testing.T) {
	type testCase struct {
		name   string
		input  string
		wantOK bool
	}

	cases := []testCase{
		{
			// Clearly an ARTIFACT/ reference.
			name:   "ARTIFACT reference",
			input:  "ARTIFACT/x(y)",
			wantOK: true,
		},
		{
			// ROOT/ reference — must return false.
			name:   "ROOT reference",
			input:  "ROOT/x(y)",
			wantOK: false,
		},
		{
			// Empty string — must return false.
			name:   "Empty string",
			input:  "",
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsArtifactRef(tc.input)
			if got != tc.wantOK {
				t.Errorf("IsArtifactRef(%q) = %v, want %v", tc.input, got, tc.wantOK)
			}
		})
	}
}

// ─── ArtifactRefParts ─────────────────────────────────────────────────────────

func TestArtifactRefParts(t *testing.T) {
	type testCase struct {
		name           string
		input          string
		wantNodePath   string
		wantArtifactID string
		wantOK         bool
	}

	cases := []testCase{
		{
			// Single-segment ARTIFACT/ with qualifier.
			name:           "ARTIFACT/x(y)",
			input:          "ARTIFACT/x(y)",
			wantNodePath:   "code-from-spec/x/_node.md",
			wantArtifactID: "y",
			wantOK:         true,
		},
		{
			// Multi-segment ARTIFACT/ with qualifier.
			name:           "ARTIFACT/x/y(z)",
			input:          "ARTIFACT/x/y(z)",
			wantNodePath:   "code-from-spec/x/y/_node.md",
			wantArtifactID: "z",
			wantOK:         true,
		},
		{
			// ARTIFACT/ without a qualifier — qualifier is required, so false.
			name:           "ARTIFACT without qualifier returns false",
			input:          "ARTIFACT/x",
			wantNodePath:   "",
			wantArtifactID: "",
			wantOK:         false,
		},
		{
			// ROOT/ reference — not an ARTIFACT/ name.
			name:           "ROOT reference returns false",
			input:          "ROOT/x(y)",
			wantNodePath:   "",
			wantArtifactID: "",
			wantOK:         false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotNodePath, gotArtifactID, gotOK := ArtifactRefParts(tc.input)
			if gotNodePath != tc.wantNodePath || gotArtifactID != tc.wantArtifactID || gotOK != tc.wantOK {
				t.Errorf(
					"ArtifactRefParts(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tc.input,
					gotNodePath, gotArtifactID, gotOK,
					tc.wantNodePath, tc.wantArtifactID, tc.wantOK,
				)
			}
		})
	}
}
