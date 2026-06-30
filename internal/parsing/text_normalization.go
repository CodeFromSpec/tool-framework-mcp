package parsing

import (
	"strings"

	"golang.org/x/text/cases"
)

func NormalizeText(rawString string) string {
	if rawString == "" {
		return ""
	}

	trimmed := trimWhitespaceNorm(rawString)
	collapsed := collapseWhitespaceNorm(trimmed)

	caser := cases.Fold()
	return caser.String(collapsed)
}

func isWhitespaceNorm(c byte) bool {
	return c == 0x20 || c == 0x09
}

func trimWhitespaceNorm(s string) string {
	start := 0
	for start < len(s) && isWhitespaceNorm(s[start]) {
		start++
	}
	end := len(s)
	for end > start && isWhitespaceNorm(s[end-1]) {
		end--
	}
	return s[start:end]
}

func collapseWhitespaceNorm(s string) string {
	var b strings.Builder
	inWhitespace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isWhitespaceNorm(c) {
			if !inWhitespace {
				b.WriteByte(0x20)
				inWhitespace = true
			}
		} else {
			b.WriteByte(c)
			inWhitespace = false
		}
	}
	return b.String()
}
