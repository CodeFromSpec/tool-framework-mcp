// code-from-spec: ROOT/golang/implementation/utils/text_normalization@8GLhf-KcYcuwErRdHvA0WCZbUaU

package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeText trims leading and trailing whitespace from raw_string,
// collapses internal runs of whitespace to a single space, and converts
// the result to lowercase using Unicode simple case folding. Returns an
// empty string when raw_string is empty or contains only whitespace.
func NormalizeText(raw_string string) string {
	if raw_string == "" {
		return ""
	}

	// Trim leading and trailing whitespace (space and horizontal tab).
	trimmed := strings.TrimFunc(raw_string, func(r rune) bool {
		return r == ' ' || r == '\t'
	})

	if trimmed == "" {
		return ""
	}

	// Collapse internal runs of whitespace (space and horizontal tab)
	// to a single space.
	var builder strings.Builder
	inWhitespace := false
	for _, r := range trimmed {
		if r == ' ' || r == '\t' {
			if !inWhitespace {
				builder.WriteRune(' ')
				inWhitespace = true
			}
		} else {
			builder.WriteRune(r)
			inWhitespace = false
		}
	}

	collapsed := builder.String()

	// Apply Unicode simple case folding.
	caser := cases.Fold()
	folded := caser.String(collapsed)

	return folded
}
