// code-from-spec: ROOT/golang/implementation/utils/text_normalization@UUHpkG5CZyXrXLRxIJQS3uCKZ1c
package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

func NormalizeText(raw_string string) string {
	trimmed := strings.TrimFunc(raw_string, func(r rune) bool {
		return r == ' ' || r == '\t'
	})

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

	caser := cases.Fold()
	return caser.String(collapsed)
}
