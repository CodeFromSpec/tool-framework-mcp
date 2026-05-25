// code-from-spec: ROOT/golang/internal/normalizename/code@PENDING
package normalizename

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeName applies the framework normalization rules to a raw heading
// or logical name qualifier text:
//  1. Trim leading and trailing whitespace (space U+0020, tab U+0009).
//  2. Collapse each run of consecutive whitespace characters to a single space.
//  3. Apply Unicode simple case folding.
func NormalizeName(raw string) string {
	// Step 1: Trim leading/trailing space and tab characters.
	s := strings.TrimFunc(raw, isSpaceOrTab)

	// Step 2: Collapse interior runs of space/tab to a single space.
	var b strings.Builder
	b.Grow(len(s))
	inWhitespace := false
	for _, r := range s {
		if isSpaceOrTab(r) {
			if !inWhitespace {
				b.WriteByte(' ')
				inWhitespace = true
			}
		} else {
			b.WriteRune(r)
			inWhitespace = false
		}
	}

	// Step 3: Unicode simple case folding.
	folder := cases.Fold()
	return folder.String(b.String())
}

// isSpaceOrTab returns true for the two whitespace characters recognized
// by the spec: space (U+0020) and horizontal tab (U+0009).
func isSpaceOrTab(r rune) bool {
	return r == ' ' || r == '\t'
}
