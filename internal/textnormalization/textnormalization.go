// code-from-spec: SPEC/golang/implementation/utils/text_normalization@NJeCREDi__PWR1VqDvpNWQNTXlc
package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

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
	return caser.String(collapsed.String())
}
