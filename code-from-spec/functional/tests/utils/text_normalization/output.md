<!-- code-from-spec: ROOT/functional/tests/utils/text_normalization@dm61A20wB8AVugzHTRSEeHtPs5Q -->

# NormalizeText — Test Specification

## Interface

```
function NormalizeText(raw_string: string) -> string
```

---

## Test Cases

### Identity

#### Already normalized

- Action: call `NormalizeText("public")`
- Expected: `"public"`

#### Single word

- Action: call `NormalizeText("Interface")`
- Expected: `"interface"`

---

### Trim

#### Leading and trailing spaces

- Action: call `NormalizeText("  Interface  ")`
- Expected: `"interface"`

#### Leading and trailing tabs

- Action: call `NormalizeText("\tInterface\t")`
- Expected: `"interface"`

#### Mixed leading whitespace

- Action: call `NormalizeText(" \t Interface \t ")`
- Expected: `"interface"`

---

### Collapse

#### Multiple spaces between words

- Action: call `NormalizeText("Testes   de   aceitacao")`
- Expected: `"testes de aceitacao"`

#### Tabs between words

- Action: call `NormalizeText("Testes\tde\taceitacao")`
- Expected: `"testes de aceitacao"`

#### Mixed whitespace between words

- Action: call `NormalizeText("Testes \t de \t aceitacao")`
- Expected: `"testes de aceitacao"`

---

### Case Folding

#### All uppercase

- Action: call `NormalizeText("PUBLIC")`
- Expected: `"public"`

#### Mixed case

- Action: call `NormalizeText("PuBLiC")`
- Expected: `"public"`

#### Unicode case folding

- Action: call `NormalizeText("TESTES DE ACEITACAO")`
- Expected: `"testes de aceitacao"`

#### German sharp s

- Action: call `NormalizeText("Strasse")`
- Expected: `"strasse"`
- Note: Unicode simple case folding maps sharp-s to `"ss"`

---

### Combined

#### Trim, collapse, and case fold together

- Action: call `NormalizeText("  TESTES   DE   ACEITACAO  ")`
- Expected: `"testes de aceitacao"`

#### Logical name qualifier style

- Action: call `NormalizeText("testes de ACEITACAO")`
- Expected: `"testes de aceitacao"`

#### Tabs and mixed case

- Action: call `NormalizeText("\tROOT/payments/fees\t")`
- Expected: `"root/payments/fees"`

---

### Edge Cases

#### Empty string

- Action: call `NormalizeText("")`
- Expected: `""`

#### Only whitespace

- Action: call `NormalizeText("   \t  ")`
- Expected: `""`

#### Non-breaking space is not whitespace

- Setup: construct a string `"hello world"` where the space between the two words is a non-breaking space (U+00A0), not a regular space
- Action: call `NormalizeText(<that string>)`
- Expected: `"hello world"` where the non-breaking space is preserved as text and not collapsed

#### Single character

- Action: call `NormalizeText("X")`
- Expected: `"x"`
