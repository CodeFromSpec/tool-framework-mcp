// code-from-spec: ROOT/golang/implementation/utils/text_normalization@QgyA3Km6OFyOGpiye-abwx2Uqbg

package textnormalization

import (
	"strings"

	"golang.org/x/text/cases"
)

// NormalizeText trims leading and trailing whitespace from raw,
// collapses any interior runs of whitespace to a single space,
// and converts all characters to their lowercase equivalents.
// Characters with Unicode folding equivalents (e.g. "ß" → "ss")
// are expanded accordingly. An empty string returns an empty string.
func NormalizeText(raw string) string {
	if raw == "" {
		return ""
	}

	// Trim leading and trailing whitespace (space and tab).
	trimmed := strings.Trim(raw, " \t")

	// Collapse interior runs of whitespace to a single space.
	// strings.FieldsFunc splits on whitespace runs; joining with a single
	// space collapses them.
	fields := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
	collapsed := strings.Join(fields, " ")

	// Apply Unicode simple case folding.
	caser := cases.Fold()
	folded := caser.String(collapsed)

	return folded
}
