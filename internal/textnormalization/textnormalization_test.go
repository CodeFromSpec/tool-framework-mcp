// code-from-spec: ROOT/golang/tests/utils/text_normalization@eQ_FYLibUd2PiPpLozCA43irUcQ
package textnormalization_test

import (
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

func TestNormalizeText(t *testing.T) {
	nbsString := "hello" + " " + "world"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already normalized", "public", "public"},
		{"single word", "Interface", "interface"},
		{"leading and trailing spaces", "  Interface  ", "interface"},
		{"leading and trailing tabs", "\tInterface\t", "interface"},
		{"mixed leading whitespace", " \t Interface \t ", "interface"},
		{"multiple spaces between words", "Testes   de   aceitacao", "testes de aceitacao"},
		{"tabs between words", "Testes\tde\taceitacao", "testes de aceitacao"},
		{"mixed whitespace between words", "Testes \t de \t aceitacao", "testes de aceitacao"},
		{"all uppercase", "PUBLIC", "public"},
		{"mixed case", "PuBLiC", "public"},
		{"unicode case folding", "TESTES DE ACEITACAO", "testes de aceitacao"},
		{"german sharp s", "Strasse", "strasse"},
		{"trim collapse and case fold together", "  TESTES   DE   ACEITACAO  ", "testes de aceitacao"},
		{"logical name qualifier style", "testes de ACEITACAO", "testes de aceitacao"},
		{"tabs and mixed case", "\tROOT/payments/fees\t", "root/payments/fees"},
		{"empty string", "", ""},
		{"only whitespace", "   \t  ", ""},
		{"non-breaking space is not whitespace", nbsString, "hello" + " " + "world"},
		{"single character", "X", "x"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := textnormalization.NormalizeText(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeText(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
