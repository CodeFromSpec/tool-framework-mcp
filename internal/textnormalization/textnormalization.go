// code-from-spec: ROOT/golang/implementation/utils/text_normalization@NaKHQX_4U3f0rRtKRBk9SkguwcE
package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

func NormalizeText(raw_string string) string {
	trimmed := strings.TrimSpace(raw_string)
	if trimmed == "" {
		return ""
	}

	fields := strings.Fields(trimmed)
	collapsed := strings.Join(fields, " ")

	caser := cases.Fold()
	return caser.String(collapsed)
}
