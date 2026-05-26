// code-from-spec: ROOT/golang/internal/normalizename/tests@aRQlm0bZ8WeFyXh_DThTQg4Sly8

// Package normalizename contains tests for the NormalizeName function.
//
// These are pure function tests — no filesystem or temp directories needed.
// All test cases are table-driven and grouped by the normalization rule(s)
// they exercise, matching the spec's section structure:
//
//   - Identity       — input already normalized
//   - Trim           — leading/trailing whitespace removal
//   - Collapse       — internal whitespace run collapsing
//   - Case folding   — Unicode simple case folding
//   - Combined       — multiple rules applied together
//   - Edge cases     — empty string, only whitespace, non-standard space, etc.
package normalizename

import (
	"testing"
)

// testCase holds a single test input/expected pair and a human-readable name.
// The prefix "test" is required by the project conventions to avoid name
// collisions with unexported identifiers in the package under test.
type testCase struct {
	name   string
	input  string
	expect string
}

// TestNormalizeName runs all spec-defined cases for NormalizeName.
// Cases are grouped into sub-tests that mirror the spec sections so that
// a failing section is immediately identifiable in test output.
func TestNormalizeName(t *testing.T) {
	// ---------------------------------------------------------------------------
	// Group: Identity — already-normalized inputs must pass through unchanged.
	// ---------------------------------------------------------------------------
	t.Run("Identity", func(t *testing.T) {
		cases := []testCase{
			{
				// A lowercase ASCII word should come back identical.
				name:   "already normalized",
				input:  "public",
				expect: "public",
			},
		}
		testRun(t, cases)
	})

	// ---------------------------------------------------------------------------
	// Group: Single word — single word with non-trivial case.
	// ---------------------------------------------------------------------------
	t.Run("SingleWord", func(t *testing.T) {
		cases := []testCase{
			{
				// Title-case word must be folded to lowercase.
				name:   "single word",
				input:  "Interface",
				expect: "interface",
			},
		}
		testRun(t, cases)
	})

	// ---------------------------------------------------------------------------
	// Group: Trim — leading and trailing whitespace (space and tab only).
	// ---------------------------------------------------------------------------
	t.Run("Trim", func(t *testing.T) {
		cases := []testCase{
			{
				// Spaces on both sides must be removed.
				name:   "leading and trailing spaces",
				input:  "  Interface  ",
				expect: "interface",
			},
			{
				// Horizontal tabs on both sides must also be removed.
				name:   "leading and trailing tabs",
				input:  "\tInterface\t",
				expect: "interface",
			},
			{
				// A mix of spaces and tabs on both sides must all be removed.
				name:   "mixed leading whitespace",
				input:  " \t Interface \t ",
				expect: "interface",
			},
		}
		testRun(t, cases)
	})

	// ---------------------------------------------------------------------------
	// Group: Collapse — internal whitespace runs become a single space.
	// ---------------------------------------------------------------------------
	t.Run("Collapse", func(t *testing.T) {
		cases := []testCase{
			{
				// Multiple spaces between words collapse to one.
				name:   "multiple spaces between words",
				input:  "Testes   de   aceitação",
				expect: "testes de aceitação",
			},
			{
				// Tabs between words also collapse to a single space.
				name:   "tabs between words",
				input:  "Testes\tde\taceitação",
				expect: "testes de aceitação",
			},
			{
				// Mixed spaces and tabs between words collapse to one space each.
				name:   "mixed whitespace between words",
				input:  "Testes \t de \t aceitação",
				expect: "testes de aceitação",
			},
		}
		testRun(t, cases)
	})

	// ---------------------------------------------------------------------------
	// Group: Case folding — Unicode simple case folding (cases.Fold()).
	// ---------------------------------------------------------------------------
	t.Run("CaseFolding", func(t *testing.T) {
		cases := []testCase{
			{
				// All-uppercase ASCII must fold to all-lowercase.
				name:   "all uppercase",
				input:  "PUBLIC",
				expect: "public",
			},
			{
				// Arbitrary mixed case must fold uniformly.
				name:   "mixed case",
				input:  "PuBLiC",
				expect: "public",
			},
			{
				// Unicode uppercase (Ã, Ç, Õ) must fold to their lowercase forms.
				name:   "unicode case folding",
				input:  "TESTES DE ACEITAÇÃO",
				expect: "testes de aceitação",
			},
			{
				// German sharp s (ß U+00DF) maps to "ss" under simple case folding.
				// This validates that the output can be longer than the input.
				name:   "german sharp s",
				input:  "Straße",
				expect: "strasse",
			},
		}
		testRun(t, cases)
	})

	// ---------------------------------------------------------------------------
	// Group: Combined — trim + collapse + case fold applied together.
	// ---------------------------------------------------------------------------
	t.Run("Combined", func(t *testing.T) {
		cases := []testCase{
			{
				// Leading/trailing spaces, internal multi-spaces, and uppercase
				// all corrected in a single call.
				name:   "trim collapse and case fold together",
				input:  "  TESTES   DE   ACEITAÇÃO  ",
				expect: "testes de aceitação",
			},
			{
				// Logical name qualifier style: lowercase with one uppercase word.
				name:   "logical name qualifier style",
				input:  "testes de ACEITAÇÃO",
				expect: "testes de aceitação",
			},
			{
				// Tabs wrapping a path-like string with uppercase; simulates a
				// raw heading that might contain a logical name qualifier.
				name:   "tabs and mixed case",
				input:  "\tROOT/payments/fees\t",
				expect: "root/payments/fees",
			},
		}
		testRun(t, cases)
	})

	// ---------------------------------------------------------------------------
	// Group: Edge cases — boundary and unusual inputs.
	// ---------------------------------------------------------------------------
	t.Run("EdgeCases", func(t *testing.T) {
		cases := []testCase{
			{
				// The empty string must pass through without modification.
				name:   "empty string",
				input:  "",
				expect: "",
			},
			{
				// A string containing only whitespace must normalize to empty.
				name:   "only whitespace",
				input:  "   \t  ",
				expect: "",
			},
			{
				// U+00A0 NO-BREAK SPACE is not in the recognized whitespace set
				// (only U+0020 and U+0009 are). It must be treated as a regular
				// character and left in place.
				name:   "non-breaking space is not whitespace",
				input:  "hello world", // "hello world" with U+00A0
				expect: "hello world",
			},
			{
				// A single uppercase character must be folded to its lowercase
				// equivalent.
				name:   "single character",
				input:  "X",
				expect: "x",
			},
		}
		testRun(t, cases)
	})
}

// testRun is a shared helper that iterates over a slice of testCase values and
// runs each one as a sub-test. The "test" prefix satisfies the project
// convention that all helper functions use this prefix to avoid collision with
// unexported identifiers in the package under test.
func testRun(t *testing.T, cases []testCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc // capture range variable for safe use in sub-test closure
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeName(tc.input)
			if got != tc.expect {
				t.Errorf(
					"NormalizeName(%q)\n  got:  %q\n  want: %q",
					tc.input, got, tc.expect,
				)
			}
		})
	}
}
