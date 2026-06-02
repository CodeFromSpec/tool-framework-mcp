// code-from-spec: ROOT/golang/tests/utils/text_normalization@FyFIsSrmBaFsbvtsvKpQPSKXjP4
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
		{
			name:  "already normalized",
			input: "public",
			want:  "public",
		},
		{
			name:  "single word",
			input: "Interface",
			want:  "interface",
		},
		{
			name:  "leading and trailing spaces",
			input: "  Interface  ",
			want:  "interface",
		},
		{
			name:  "leading and trailing tabs",
			input: "\tInterface\t",
			want:  "interface",
		},
		{
			name:  "mixed leading whitespace",
			input: " \t Interface \t ",
			want:  "interface",
		},
		{
			name:  "multiple spaces between words",
			input: "Testes   de   aceitacao",
			want:  "testes de aceitacao",
		},
		{
			name:  "tabs between words",
			input: "Testes\tde\taceitacao",
			want:  "testes de aceitacao",
		},
		{
			name:  "mixed whitespace between words",
			input: "Testes \t de \t aceitacao",
			want:  "testes de aceitacao",
		},
		{
			name:  "all uppercase",
			input: "PUBLIC",
			want:  "public",
		},
		{
			name:  "mixed case",
			input: "PuBLiC",
			want:  "public",
		},
		{
			name:  "unicode case folding",
			input: "TESTES DE ACEITACAO",
			want:  "testes de aceitacao",
		},
		{
			name:  "german sharp s",
			input: "Strasse",
			want:  "strasse",
		},
		{
			name:  "trim collapse and case fold together",
			input: "  TESTES   DE   ACEITACAO  ",
			want:  "testes de aceitacao",
		},
		{
			name:  "logical name qualifier style",
			input: "testes de ACEITACAO",
			want:  "testes de aceitacao",
		},
		{
			name:  "tabs and mixed case",
			input: "\tROOT/payments/fees\t",
			want:  "root/payments/fees",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only whitespace",
			input: "   \t  ",
			want:  "",
		},
		{
			name:  "non-breaking space is not whitespace",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "single character",
			input: "X",
			want:  "x",
		},
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
