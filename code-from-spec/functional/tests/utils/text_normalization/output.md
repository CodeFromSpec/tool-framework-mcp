<!-- code-from-spec: ROOT/functional/tests/utils/text_normalization@F8imG1q43CONfOdM_zEVTF8SD48 -->

# Test Specification: NormalizeText

## Function

```
NormalizeText(raw_string: string) -> string
```

---

## Test Cases

### Identity

#### Already normalized

- Setup: none
- Action: call `NormalizeText("public")`
- Expected: `"public"`

#### Single word

- Setup: none
- Action: call `NormalizeText("Interface")`
- Expected: `"interface"`

---

### Trim

#### Leading and trailing spaces

- Setup: none
- Action: call `NormalizeText("  Interface  ")`
- Expected: `"interface"`

#### Leading and trailing tabs

- Setup: none
- Action: call `NormalizeText("\tInterface\t")`
- Expected: `"interface"`

#### Mixed leading whitespace

- Setup: none
- Action: call `NormalizeText(" \t Interface \t ")`
- Expected: `"interface"`

---

### Collapse

#### Multiple spaces between words

- Setup: none
- Action: call `NormalizeText("Testes   de   aceitacao")`
- Expected: `"testes de aceitacao"`

#### Tabs between words

- Setup: none
- Action: call `NormalizeText("Testes\tde\taceitacao")`
- Expected: `"testes de aceitacao"`

#### Mixed whitespace between words

- Setup: none
- Action: call `NormalizeText("Testes \t de \t aceitacao")`
- Expected: `"testes de aceitacao"`

---

### Case Folding

#### All uppercase

- Setup: none
- Action: call `NormalizeText("PUBLIC")`
- Expected: `"public"`

#### Mixed case

- Setup: none
- Action: call `NormalizeText("PuBLiC")`
- Expected: `"public"`

#### Unicode case folding

- Setup: none
- Action: call `NormalizeText("TESTES DE ACEITACAO")`
- Expected: `"testes de aceitacao"`

#### German sharp s

- Setup: none
- Action: call `NormalizeText("Strasse")`
- Expected: `"strasse"`
- Note: Unicode simple case folding maps sharp-s to `"ss"`

---

### Combined

#### Trim, collapse, and case fold together

- Setup: none
- Action: call `NormalizeText("  TESTES   DE   ACEITACAO  ")`
- Expected: `"testes de aceitacao"`

#### Logical name qualifier style

- Setup: none
- Action: call `NormalizeText("testes de ACEITACAO")`
- Expected: `"testes de aceitacao"`

#### Tabs and mixed case

- Setup: none
- Action: call `NormalizeText("\tROOT/payments/fees\t")`
- Expected: `"root/payments/fees"`

---

### Edge Cases

#### Empty string

- Setup: none
- Action: call `NormalizeText("")`
- Expected: `""`

#### Only whitespace

- Setup: none
- Action: call `NormalizeText("   \t  ")`
- Expected: `""`

#### Non-breaking space is not whitespace

- Setup: none
- Action: call `NormalizeText("hello world")` (the space between the words is a non-breaking space U+00A0)
- Expected: `"hello world"` (non-breaking space is treated as text, not collapsed or trimmed)

#### Single character

- Setup: none
- Action: call `NormalizeText("X")`
- Expected: `"x"`
