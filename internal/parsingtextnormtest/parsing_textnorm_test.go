// code-from-spec: SPEC/golang/tests/parsing/text_normalization@W8CVxfcd09e-hLnmUBMNDGDkX5E
package parsingtextnormtest_test

import (
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "already normalized",
			input:  "public",
			expect: "public",
		},
		{
			name:   "single word",
			input:  "Interface",
			expect: "interface",
		},
		{
			name:   "leading and trailing spaces",
			input:  "  Interface  ",
			expect: "interface",
		},
		{
			name:   "leading and trailing tabs",
			input:  "\tInterface\t",
			expect: "interface",
		},
		{
			name:   "mixed leading whitespace",
			input:  " \t Interface \t ",
			expect: "interface",
		},
		{
			name:   "multiple spaces between words",
			input:  "Testes   de   aceitacao",
			expect: "testes de aceitacao",
		},
		{
			name:   "tabs between words",
			input:  "Testes\tde\taceitacao",
			expect: "testes de aceitacao",
		},
		{
			name:   "mixed whitespace between words",
			input:  "Testes \t de \t aceitacao",
			expect: "testes de aceitacao",
		},
		{
			name:   "all uppercase",
			input:  "PUBLIC",
			expect: "public",
		},
		{
			name:   "mixed case",
			input:  "PuBLiC",
			expect: "public",
		},
		{
			name:   "unicode case folding",
			input:  "TESTES DE ACEITACAO",
			expect: "testes de aceitacao",
		},
		{
			name:   "german sharp s",
			input:  "Straße",
			expect: "strasse",
		},
		{
			name:   "trim collapse and case fold together",
			input:  "  TESTES   DE   ACEITACAO  ",
			expect: "testes de aceitacao",
		},
		{
			name:   "logical name qualifier style",
			input:  "testes de ACEITACAO",
			expect: "testes de aceitacao",
		},
		{
			name:   "tabs and mixed case",
			input:  "\tROOT/payments/fees\t",
			expect: "root/payments/fees",
		},
		{
			name:   "empty string",
			input:  "",
			expect: "",
		},
		{
			name:   "only whitespace",
			input:  "   \t  ",
			expect: "",
		},
		{
			name:   "non-breaking space is not whitespace",
			input:  "hello" + " " + "world",
			expect: "hello" + " " + "world",
		},
		{
			name:   "single character",
			input:  "X",
			expect: "x",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parsing.NormalizeText(tc.input)
			if got != tc.expect {
				t.Errorf("NormalizeText(%q) = %q, want %q", tc.input, got, tc.expect)
			}
		})
	}
}
