// code-from-spec: ROOT/golang/implementation/utils/text_normalization@w8aFhnIAtczjcnBRMFEOwqjIuCY

package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeText trims leading and trailing whitespace from raw,
// converts it to lowercase, and expands Unicode characters
// (e.g., "Straße" → "strasse"). Multiple internal spaces are
// collapsed to a single space. Returns an empty string unchanged.
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
	folded := caser.String(collapsed)

	return folded
}
