---
depends_on:
  - ROOT/functional/logic/utils/text_normalization(interface)
output: code-from-spec/functional/tests/utils/text_normalization/output.md
---

# ROOT/functional/tests/utils/text_normalization

Test cases for the text normalization component.

# Public

## Test cases

### Identity

#### Already normalized

Input: "public". Expect: "public".

#### Single word

Input: "Interface". Expect: "interface".

### Trim

#### Leading and trailing spaces

Input: "  Interface  ". Expect: "interface".

#### Leading and trailing tabs

Input: "\tInterface\t". Expect: "interface".

#### Mixed leading whitespace

Input: " \t Interface \t ". Expect: "interface".

### Collapse

#### Multiple spaces between words

Input: "Testes   de   aceitacao". Expect:
"testes de aceitacao".

#### Tabs between words

Input: "Testes\tde\taceitacao". Expect:
"testes de aceitacao".

#### Mixed whitespace between words

Input: "Testes \t de \t aceitacao". Expect:
"testes de aceitacao".

### Case folding

#### All uppercase

Input: "PUBLIC". Expect: "public".

#### Mixed case

Input: "PuBLiC". Expect: "public".

#### Unicode case folding

Input: "TESTES DE ACEITACAO". Expect: "testes de aceitacao".

#### German sharp s

Input: "Strasse". Expect: "strasse" (Unicode simple case
folding maps sharp-s to "ss").

### Combined

#### Trim, collapse, and case fold together

Input: "  TESTES   DE   ACEITACAO  ". Expect:
"testes de aceitacao".

#### Logical name qualifier style

Input: "testes de ACEITACAO". Expect: "testes de aceitacao".

#### Tabs and mixed case

Input: "\tROOT/payments/fees\t". Expect:
"root/payments/fees".

### Edge cases

#### Empty string

Input: "". Expect: "".

#### Only whitespace

Input: "   \t  ". Expect: "".

#### Non-breaking space is not whitespace

Input: "hello world" (contains non-breaking space).
Expect: "hello world" (non-breaking space treated as
text, not collapsed).

#### Single character

Input: "X". Expect: "x".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `NormalizeText`.
