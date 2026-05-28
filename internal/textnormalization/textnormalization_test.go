// code-from-spec: ROOT/golang/tests/utils/text_normalization@zOKpYUbXY2YXD6-OMezM3IjoaDQ

package textnormalization_test

import (
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/textnormalization"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Group: Identity
		{name: "TC-IDENTITY-01 already normalized", input: "public", want: "public"},
		{name: "TC-IDENTITY-02 single word with initial capital", input: "Interface", want: "interface"},

		// Group: Trim
		{name: "TC-TRIM-01 leading and trailing spaces", input: "  Interface  ", want: "interface"},
		{name: "TC-TRIM-02 leading and trailing tabs", input: "\tInterface\t", want: "interface"},
		{name: "TC-TRIM-03 mixed leading and trailing whitespace", input: " \t Interface \t ", want: "interface"},

		// Group: Collapse
		{name: "TC-COLLAPSE-01 multiple spaces between words", input: "Testes   de   aceitacao", want: "testes de aceitacao"},
		{name: "TC-COLLAPSE-02 tabs between words", input: "Testes\tde\taceitacao", want: "testes de aceitacao"},
		{name: "TC-COLLAPSE-03 mixed whitespace between words", input: "Testes \t de \t aceitacao", want: "testes de aceitacao"},

		// Group: Case Folding
		{name: "TC-CASE-01 all uppercase", input: "PUBLIC", want: "public"},
		{name: "TC-CASE-02 mixed case", input: "PuBLiC", want: "public"},
		{name: "TC-CASE-03 unicode uppercase", input: "TESTES DE ACEITACAO", want: "testes de aceitacao"},
		{name: "TC-CASE-04 german sharp-s already decomposed", input: "Strasse", want: "strasse"},

		// Group: Combined
		{name: "TC-COMBINED-01 trim collapse and case fold together", input: "  TESTES   DE   ACEITACAO  ", want: "testes de aceitacao"},
		{name: "TC-COMBINED-02 logical name qualifier style", input: "testes de ACEITACAO", want: "testes de aceitacao"},
		{name: "TC-COMBINED-03 tabs and mixed case with path-like content", input: "\tROOT/payments/fees\t", want: "root/payments/fees"},

		// Group: Edge Cases
		{name: "TC-EDGE-01 empty string", input: "", want: ""},
		{name: "TC-EDGE-02 only whitespace", input: "   \t  ", want: ""},
		{name: "TC-EDGE-03 non-breaking space is not treated as whitespace", input: "hello world", want: "hello world"},
		{name: "TC-EDGE-04 single character", input: "X", want: "x"},
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
