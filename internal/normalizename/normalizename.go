// code-from-spec: ROOT/golang/internal/normalizename/code@Cp4o1Qg03pIPOvnWGBKWd0Zqd-s

// Package normalizename provides the NormalizeName function, which applies
// the framework's canonical normalization rules to a raw heading or logical
// name qualifier string.
//
// Normalization pipeline (applied in order):
//  1. Trim leading and trailing whitespace (U+0020 and U+0009 only).
//  2. Collapse internal runs of whitespace (U+0020 and U+0009) to a single
//     space (U+0020).
//  3. Apply Unicode simple case folding via golang.org/x/text/cases.Fold().
//
// The function is a pure function: no I/O, no external state, no errors.
package normalizename

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeName applies the framework normalization rules to raw and returns
// the normalized string. It is safe to call with any input including empty
// strings.
//
// Whitespace characters recognised by this function are space (U+0020) and
// horizontal tab (U+0009) only. Other Unicode whitespace (e.g. U+00A0
// no-break space, U+2003 em space) is NOT treated as whitespace here.
func NormalizeName(raw string) string {
	// Step 1 – short-circuit on empty input; avoids unnecessary allocations.
	if raw == "" {
		return ""
	}

	// Step 2 – trim leading and trailing whitespace (space and tab only).
	// strings.Trim with the explicit cutset is used instead of
	// strings.TrimSpace so that only U+0020 and U+0009 are stripped.
	trimmed := strings.Trim(raw, " \t")

	// Step 3 – collapse internal runs of space/tab into a single space.
	// We walk through the trimmed string once, copying non-whitespace
	// characters directly and replacing each whitespace run with exactly
	// one U+0020. This avoids a regexp dependency and keeps the logic
	// explicit and easy to verify.
	var b strings.Builder
	b.Grow(len(trimmed)) // pre-allocate; output is never longer than input
	inSpace := false
	for _, r := range trimmed {
		if r == ' ' || r == '\t' {
			// We are inside a whitespace run. Emit a single space the first
			// time we enter the run; skip subsequent whitespace characters.
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
		} else {
			// Non-whitespace character: copy it through and reset the flag.
			b.WriteRune(r)
			inSpace = false
		}
	}
	collapsed := b.String()

	// Step 4 – apply Unicode simple case folding.
	// cases.Fold() returns a Caser that implements the Unicode Simple_Case_Folding
	// mapping (CaseFolding.txt "S" and "C" entries). It is applied code point
	// by code point, which means the output length may differ from the input
	// length (e.g. "ß" U+00DF → "ss").
	caser := cases.Fold()
	folded := caser.String(collapsed)

	return folded
}
