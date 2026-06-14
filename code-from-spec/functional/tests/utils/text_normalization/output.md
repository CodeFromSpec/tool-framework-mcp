<!-- code-from-spec: ROOT/functional/tests/utils/text_normalization@yJGBAKMagC9RoPC4ZQrOcXrtkl8 -->

## Test suite: NormalizeText

---

### Identity

#### Already normalized

Setup: none.
Action: call NormalizeText("public").
Expected: returns "public".

#### Single word

Setup: none.
Action: call NormalizeText("Interface").
Expected: returns "interface".

---

### Trim

#### Leading and trailing spaces

Setup: none.
Action: call NormalizeText("  Interface  ").
Expected: returns "interface".

#### Leading and trailing tabs

Setup: none.
Action: call NormalizeText("\tInterface\t").
Expected: returns "interface".

#### Mixed leading whitespace

Setup: none.
Action: call NormalizeText(" \t Interface \t ").
Expected: returns "interface".

---

### Collapse

#### Multiple spaces between words

Setup: none.
Action: call NormalizeText("Testes   de   aceitacao").
Expected: returns "testes de aceitacao".

#### Tabs between words

Setup: none.
Action: call NormalizeText("Testes\tde\taceitacao").
Expected: returns "testes de aceitacao".

#### Mixed whitespace between words

Setup: none.
Action: call NormalizeText("Testes \t de \t aceitacao").
Expected: returns "testes de aceitacao".

---

### Case folding

#### All uppercase

Setup: none.
Action: call NormalizeText("PUBLIC").
Expected: returns "public".

#### Mixed case

Setup: none.
Action: call NormalizeText("PuBLiC").
Expected: returns "public".

#### Unicode case folding

Setup: none.
Action: call NormalizeText("TESTES DE ACEITACAO").
Expected: returns "testes de aceitacao".

#### German sharp s

Setup: none.
Action: call NormalizeText("Strasse").
Expected: returns "strasse".

---

### Combined

#### Trim, collapse, and case fold together

Setup: none.
Action: call NormalizeText("  TESTES   DE   ACEITACAO  ").
Expected: returns "testes de aceitacao".

#### Logical name qualifier style

Setup: none.
Action: call NormalizeText("testes de ACEITACAO").
Expected: returns "testes de aceitacao".

#### Tabs and mixed case

Setup: none.
Action: call NormalizeText("\tROOT/payments/fees\t").
Expected: returns "root/payments/fees".

---

### Edge cases

#### Empty string

Setup: none.
Action: call NormalizeText("").
Expected: returns "".

#### Only whitespace

Setup: none.
Action: call NormalizeText("   \t  ").
Expected: returns "".

#### Non-breaking space is not whitespace

Setup: input string contains the word "hello", followed by
a non-breaking space character (U+00A0), followed by the
word "world".
Action: call NormalizeText with that input string.
Expected: returns a string where the non-breaking space is
preserved as text, not collapsed or trimmed — the result
is "hello" + non-breaking space + "world", all lowercased.

#### Single character

Setup: none.
Action: call NormalizeText("X").
Expected: returns "x".
