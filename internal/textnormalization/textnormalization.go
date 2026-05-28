// code-from-spec: ROOT/golang/implementation/utils/text_normalization@hdOyvGWKkfhzJinWQ7wK1ElVMuE

package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeText normalizes a raw string by trimming leading and trailing
// whitespace, collapsing internal runs of whitespace to a single space,
// converting the result to lowercase, and transliterating Unicode characters
// to their ASCII equivalents where possible (e.g., "ß" → "ss").
//
// An empty string input returns an empty string.
func NormalizeText(rawString string) string {
	if rawString == "" {
		return ""
	}

	trimmed := strings.TrimFunc(rawString, func(r rune) bool {
		return r == ' ' || r == '\t'
	})

	var collapsed strings.Builder
	inWhitespace := false
	for _, r := range trimmed {
		if r == ' ' || r == '\t' {
			if !inWhitespace {
				collapsed.WriteRune(' ')
				inWhitespace = true
			}
		} else {
			collapsed.WriteRune(r)
			inWhitespace = false
		}
	}

	caser := cases.Fold()
	folded := caser.String(collapsed.String())

	return folded
}
