// code-from-spec: ROOT/golang/tests/utils/text_normalization@KK8woURiyCRqeez-q7UUcqBJAwk
package textnormalization_test

import (
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Identity
		{name: "TC-01 already normalized", input: "public", want: "public"},
		{name: "TC-02 single word", input: "Interface", want: "interface"},

		// Trim
		{name: "TC-03 leading and trailing spaces", input: "  Interface  ", want: "interface"},
		{name: "TC-04 leading and trailing tabs", input: "\tInterface\t", want: "interface"},
		{name: "TC-05 mixed leading whitespace", input: " \t Interface \t ", want: "interface"},

		// Collapse
		{name: "TC-06 multiple spaces between words", input: "Testes   de   aceitacao", want: "testes de aceitacao"},
		{name: "TC-07 tabs between words", input: "Testes\tde\taceitacao", want: "testes de aceitacao"},
		{name: "TC-08 mixed whitespace between words", input: "Testes \t de \t aceitacao", want: "testes de aceitacao"},

		// Case Folding
		{name: "TC-09 all uppercase", input: "PUBLIC", want: "public"},
		{name: "TC-10 mixed case", input: "PuBLiC", want: "public"},
		{name: "TC-11 unicode case folding", input: "TESTES DE ACEITACAO", want: "testes de aceitacao"},
		{name: "TC-12 german sharp s", input: "Strasse", want: "strasse"},

		// Combined
		{name: "TC-13 trim collapse and case fold together", input: "  TESTES   DE   ACEITACAO  ", want: "testes de aceitacao"},
		{name: "TC-14 logical name qualifier style", input: "testes de ACEITACAO", want: "testes de aceitacao"},
		{name: "TC-15 tabs and mixed case", input: "\tROOT/payments/fees\t", want: "root/payments/fees"},

		// Edge Cases
		{name: "TC-16 empty string", input: "", want: ""},
		{name: "TC-17 only whitespace", input: "   \t  ", want: ""},
		{name: "TC-18 non-breaking space is not whitespace", input: "hello world", want: "hello world"},
		{name: "TC-19 single character", input: "X", want: "x"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := textnormalization.NormalizeText(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeText(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
