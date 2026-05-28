// code-from-spec: ROOT/golang/implementation/internal/textnormalization/code@EJ4-KMjk4QQT-OaJVM-M8mpZL-0

package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeText applies the framework normalization rules to a raw heading
// or qualifier text:
//  1. Trim leading and trailing whitespace (space U+0020, tab U+0009).
//  2. Collapse each sequence of one or more whitespace characters to a single space.
//  3. Apply Unicode simple case folding.
func NormalizeText(raw string) string {
	if raw == "" {
		return ""
	}

	// Step 1: Trim leading and trailing whitespace (space and tab).
	trimmed := strings.TrimFunc(raw, isWhitespace)

	// Step 2: Collapse runs of whitespace to a single space.
	var builder strings.Builder
	inWhitespace := false
	for _, ch := range trimmed {
		if isWhitespace(ch) {
			if !inWhitespace {
				builder.WriteRune(' ')
				inWhitespace = true
			}
		} else {
			builder.WriteRune(ch)
			inWhitespace = false
		}
	}
	collapsed := builder.String()

	// Step 3: Apply Unicode simple case folding.
	caser := cases.Fold()
	folded := caser.String(collapsed)

	return folded
}

// isWhitespace reports whether r is a whitespace character as defined by
// the normalization rules: space (U+0020) or horizontal tab (U+0009).
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t'
}
