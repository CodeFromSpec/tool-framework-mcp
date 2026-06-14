// code-from-spec: ROOT/golang/implementation/utils/text_normalization@BKW7Ca1MhRDi-6_hThdG73eOYpc
package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

func NormalizeText(rawString string) string {
	trimmed := strings.Trim(rawString, " \t")

	var builder strings.Builder
	inWhitespace := false

	for _, ch := range trimmed {
		if ch == ' ' || ch == '\t' {
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

	caser := cases.Fold()
	return caser.String(collapsed)
}
