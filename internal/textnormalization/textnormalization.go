// code-from-spec: ROOT/golang/implementation/utils/text_normalization@XGOo0KCIR6WFlANIGYXpClOzkgM
package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

func NormalizeText(raw string) string {
	if raw == "" {
		return ""
	}

	trimmed := strings.Trim(raw, " \t")

	fields := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
	collapsed := strings.Join(fields, " ")

	caser := cases.Fold()
	return caser.String(collapsed)
}
