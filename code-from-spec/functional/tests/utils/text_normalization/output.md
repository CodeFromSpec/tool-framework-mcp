<!-- code-from-spec: ROOT/functional/tests/utils/text_normalization@aLvx1bSAlVdgZhC4de-eVqwPx30 -->

# Test Specification: NormalizeText

## Interface

```
function NormalizeText(raw_string) -> string
```

Normalizes a raw string by trimming leading and trailing whitespace,
collapsing internal whitespace runs to a single space, and applying
Unicode simple case folding (lowercase). Non-breaking spaces are
treated as regular text characters, not as whitespace.

---

## Test Cases

### Group: Identity

#### TC-IDENTITY-01 — Already normalized

- Setup: none
- Action: call `NormalizeText("public")`
- Expected outcome: `"public"`

#### TC-IDENTITY-02 — Single word with initial capital

- Setup: none
- Action: call `NormalizeText("Interface")`
- Expected outcome: `"interface"`

---

### Group: Trim

#### TC-TRIM-01 — Leading and trailing spaces

- Setup: none
- Action: call `NormalizeText("  Interface  ")`
- Expected outcome: `"interface"`

#### TC-TRIM-02 — Leading and trailing tabs

- Setup: none
- Action: call `NormalizeText("\tInterface\t")`
- Expected outcome: `"interface"`

#### TC-TRIM-03 — Mixed leading and trailing whitespace

- Setup: none
- Action: call `NormalizeText(" \t Interface \t ")`
- Expected outcome: `"interface"`

---

### Group: Collapse

#### TC-COLLAPSE-01 — Multiple spaces between words

- Setup: none
- Action: call `NormalizeText("Testes   de   aceitacao")`
- Expected outcome: `"testes de aceitacao"`

#### TC-COLLAPSE-02 — Tabs between words

- Setup: none
- Action: call `NormalizeText("Testes\tde\taceitacao")`
- Expected outcome: `"testes de aceitacao"`

#### TC-COLLAPSE-03 — Mixed whitespace between words

- Setup: none
- Action: call `NormalizeText("Testes \t de \t aceitacao")`
- Expected outcome: `"testes de aceitacao"`

---

### Group: Case Folding

#### TC-CASE-01 — All uppercase

- Setup: none
- Action: call `NormalizeText("PUBLIC")`
- Expected outcome: `"public"`

#### TC-CASE-02 — Mixed case

- Setup: none
- Action: call `NormalizeText("PuBLiC")`
- Expected outcome: `"public"`

#### TC-CASE-03 — Unicode uppercase

- Setup: none
- Action: call `NormalizeText("TESTES DE ACEITACAO")`
- Expected outcome: `"testes de aceitacao"`

#### TC-CASE-04 — German sharp-s (ß variant already decomposed)

- Setup: none
- Action: call `NormalizeText("Strasse")`
- Expected outcome: `"strasse"`
  Note: Unicode simple case folding maps the sharp-s character (ß) to
  `"ss"`. The input here uses the already-decomposed form `"Strasse"`,
  so the result is simply the lowercased string `"strasse"`.

---

### Group: Combined

#### TC-COMBINED-01 — Trim, collapse, and case fold together

- Setup: none
- Action: call `NormalizeText("  TESTES   DE   ACEITACAO  ")`
- Expected outcome: `"testes de aceitacao"`

#### TC-COMBINED-02 — Logical name qualifier style

- Setup: none
- Action: call `NormalizeText("testes de ACEITACAO")`
- Expected outcome: `"testes de aceitacao"`

#### TC-COMBINED-03 — Tabs and mixed case with path-like content

- Setup: none
- Action: call `NormalizeText("\tROOT/payments/fees\t")`
- Expected outcome: `"root/payments/fees"`

---

### Group: Edge Cases

#### TC-EDGE-01 — Empty string

- Setup: none
- Action: call `NormalizeText("")`
- Expected outcome: `""`

#### TC-EDGE-02 — Only whitespace

- Setup: none
- Action: call `NormalizeText("   \t  ")`
- Expected outcome: `""`

#### TC-EDGE-03 — Non-breaking space is not treated as whitespace

- Setup: The input string contains a non-breaking space (U+00A0)
  between `"hello"` and `"world"`, written as `"hello world"`.
- Action: call `NormalizeText("hello world")`
- Expected outcome: `"hello world"`
  Note: The non-breaking space must not be collapsed or trimmed; it
  is treated as a regular text character. The result is still
  lowercased but the spacing character is preserved as-is.

#### TC-EDGE-04 — Single character

- Setup: none
- Action: call `NormalizeText("X")`
- Expected outcome: `"x"`
