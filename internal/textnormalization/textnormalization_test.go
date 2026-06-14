// code-from-spec: ROOT/golang/tests/utils/text_normalization@llNgJxzuUWJJRoQCoRjfHAzqMw4
package textnormalization_test

import (
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized",
			input:    "public",
			expected: "public",
		},
		{
			name:     "single word",
			input:    "Interface",
			expected: "interface",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  Interface  ",
			expected: "interface",
		},
		{
			name:     "leading and trailing tabs",
			input:    "\tInterface\t",
			expected: "interface",
		},
		{
			name:     "mixed leading whitespace",
			input:    " \t Interface \t ",
			expected: "interface",
		},
		{
			name:     "multiple spaces between words",
			input:    "Testes   de   aceitacao",
			expected: "testes de aceitacao",
		},
		{
			name:     "tabs between words",
			input:    "Testes\tde\taceitacao",
			expected: "testes de aceitacao",
		},
		{
			name:     "mixed whitespace between words",
			input:    "Testes \t de \t aceitacao",
			expected: "testes de aceitacao",
		},
		{
			name:     "all uppercase",
			input:    "PUBLIC",
			expected: "public",
		},
		{
			name:     "mixed case",
			input:    "PuBLiC",
			expected: "public",
		},
		{
			name:     "unicode case folding",
			input:    "TESTES DE ACEITACAO",
			expected: "testes de aceitacao",
		},
		{
			name:     "german sharp s",
			input:    "Strasse",
			expected: "strasse",
		},
		{
			name:     "trim collapse and case fold together",
			input:    "  TESTES   DE   ACEITACAO  ",
			expected: "testes de aceitacao",
		},
		{
			name:     "logical name qualifier style",
			input:    "testes de ACEITACAO",
			expected: "testes de aceitacao",
		},
		{
			name:     "tabs and mixed case",
			input:    "\tROOT/payments/fees\t",
			expected: "root/payments/fees",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t  ",
			expected: "",
		},
		{
			name:     "non-breaking space is not whitespace",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "single character",
			input:    "X",
			expected: "x",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := textnormalization.NormalizeText(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeText(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
