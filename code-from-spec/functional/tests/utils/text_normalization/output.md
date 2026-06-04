<!-- code-from-spec: ROOT/functional/tests/utils/text_normalization@Kl0fpKwpdky-twBLNB2Y40EGPiY -->

## Identity

### Already normalized

Actions: call NormalizeText("public").
Expected: returns "public".

### Single word

Actions: call NormalizeText("Interface").
Expected: returns "interface".

## Trim

### Leading and trailing spaces

Actions: call NormalizeText("  Interface  ").
Expected: returns "interface".

### Leading and trailing tabs

Actions: call NormalizeText("\tInterface\t").
Expected: returns "interface".

### Mixed leading whitespace

Actions: call NormalizeText(" \t Interface \t ").
Expected: returns "interface".

## Collapse

### Multiple spaces between words

Actions: call NormalizeText("Testes   de   aceitacao").
Expected: returns "testes de aceitacao".

### Tabs between words

Actions: call NormalizeText("Testes\tde\taceitacao").
Expected: returns "testes de aceitacao".

### Mixed whitespace between words

Actions: call NormalizeText("Testes \t de \t aceitacao").
Expected: returns "testes de aceitacao".

## Case folding

### All uppercase

Actions: call NormalizeText("PUBLIC").
Expected: returns "public".

### Mixed case

Actions: call NormalizeText("PuBLiC").
Expected: returns "public".

### Unicode case folding

Actions: call NormalizeText("TESTES DE ACEITACAO").
Expected: returns "testes de aceitacao".

### German sharp s

Actions: call NormalizeText("Strasse").
Expected: returns "strasse".

## Combined

### Trim, collapse, and case fold together

Actions: call NormalizeText("  TESTES   DE   ACEITACAO  ").
Expected: returns "testes de aceitacao".

### Logical name qualifier style

Actions: call NormalizeText("testes de ACEITACAO").
Expected: returns "testes de aceitacao".

### Tabs and mixed case

Actions: call NormalizeText("\tROOT/payments/fees\t").
Expected: returns "root/payments/fees".

## Edge cases

### Empty string

Actions: call NormalizeText("").
Expected: returns "".

### Only whitespace

Actions: call NormalizeText("   \t  ").
Expected: returns "".

### Non-breaking space is not whitespace

Setup: construct a string "hello" + non-breaking space (U+00A0) + "world".
Actions: call NormalizeText with that string.
Expected: returns "hello" + non-breaking space + "world" (non-breaking space is preserved as text, not collapsed or trimmed).

### Single character

Actions: call NormalizeText("X").
Expected: returns "x".
