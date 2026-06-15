<!-- code-from-spec: SPEC/functional/tests/utils/text_normalization@F3fqNWTJ9XXUdkqLwg7EfFC47tU -->

## Test suite: NormalizeText

---

### Identity

#### Already normalized

Action: call NormalizeText("public").
Expected: returns "public".

#### Single word

Action: call NormalizeText("Interface").
Expected: returns "interface".

---

### Trim

#### Leading and trailing spaces

Action: call NormalizeText("  Interface  ").
Expected: returns "interface".

#### Leading and trailing tabs

Action: call NormalizeText("\tInterface\t").
Expected: returns "interface".

#### Mixed leading whitespace

Action: call NormalizeText(" \t Interface \t ").
Expected: returns "interface".

---

### Collapse

#### Multiple spaces between words

Action: call NormalizeText("Testes   de   aceitacao").
Expected: returns "testes de aceitacao".

#### Tabs between words

Action: call NormalizeText("Testes\tde\taceitacao").
Expected: returns "testes de aceitacao".

#### Mixed whitespace between words

Action: call NormalizeText("Testes \t de \t aceitacao").
Expected: returns "testes de aceitacao".

---

### Case folding

#### All uppercase

Action: call NormalizeText("PUBLIC").
Expected: returns "public".

#### Mixed case

Action: call NormalizeText("PuBLiC").
Expected: returns "public".

#### Unicode case folding

Action: call NormalizeText("TESTES DE ACEITACAO").
Expected: returns "testes de aceitacao".

#### German sharp s

Action: call NormalizeText("Strasse").
Expected: returns "strasse".

---

### Combined

#### Trim, collapse, and case fold together

Action: call NormalizeText("  TESTES   DE   ACEITACAO  ").
Expected: returns "testes de aceitacao".

#### Logical name qualifier style

Action: call NormalizeText("testes de ACEITACAO").
Expected: returns "testes de aceitacao".

#### Tabs and mixed case

Action: call NormalizeText("\tROOT/payments/fees\t").
Expected: returns "root/payments/fees".

---

### Edge cases

#### Empty string

Action: call NormalizeText("").
Expected: returns "".

#### Only whitespace

Action: call NormalizeText("   \t  ").
Expected: returns "".

#### Non-breaking space is not whitespace

Setup: construct an input string consisting of "hello", a non-breaking space character (U+00A0), and "world" — no regular spaces.
Action: call NormalizeText with that string.
Expected: returns "hello world" — the non-breaking space is preserved as text, not collapsed or removed.

#### Single character

Action: call NormalizeText("X").
Expected: returns "x".
