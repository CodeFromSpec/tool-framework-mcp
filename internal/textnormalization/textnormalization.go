// code-from-spec: ROOT/golang/implementation/utils/text_normalization@iKxTjSYfHkQgQ2j5X1cFJ48VeQw
package textnormalization

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
)

func NormalizeText(raw string) string {
	trimmed := strings.TrimFunc(raw, unicode.IsSpace)
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	inSpace := false
	for _, r := range trimmed {
		if unicode.IsSpace(r) {
			if !inSpace {
				builder.WriteRune(' ')
				inSpace = true
			}
		} else {
			builder.WriteRune(r)
			inSpace = false
		}
	}
	collapsed := builder.String()

	caser := cases.Fold()
	return caser.String(collapsed)
}
