// code-from-spec: ROOT/golang/tests/internal/textnormalization@TUxFx11Hpc34wzmg5gmRDq_re2Q

package textnormalization_test

import (
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/textnormalization"
)

func TestNormalizeText(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  string
	}

	tests := []testCase{
		// identity
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

		// trim
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

		// collapse
		{
			name:  "multiple spaces between words",
			input: "Testes   de   aceitação",
			want:  "testes de aceitação",
		},
		{
			name:  "tabs between words",
			input: "Testes\tde\taceitação",
			want:  "testes de aceitação",
		},
		{
			name:  "mixed whitespace between words",
			input: "Testes \t de \t aceitação",
			want:  "testes de aceitação",
		},

		// case folding
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
			input: "TESTES DE ACEITAÇÃO",
			want:  "testes de aceitação",
		},
		{
			name:  "german sharp s",
			input: "Straße",
			want:  "strasse",
		},

		// combined
		{
			name:  "trim collapse and case fold together",
			input: "  TESTES   DE   ACEITAÇÃO  ",
			want:  "testes de aceitação",
		},
		{
			name:  "logical name qualifier style",
			input: "testes de ACEITAÇÃO",
			want:  "testes de aceitação",
		},
		{
			name:  "tabs and mixed case",
			input: "\tROOT/payments/fees\t",
			want:  "root/payments/fees",
		},

		// edge cases
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
				t.Errorf("NormalizeText(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
