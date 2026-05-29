<!-- code-from-spec: ROOT/functional/tests/utils/text_normalization@aLvx1bSAlVdgZhC4de-eVqwPx30 -->

# Test Specification: NormalizeText

## Function Under Test

```
NormalizeText(raw_string: string) -> string
```

---

## Test Cases

### Identity

---

#### TC-01: Already normalized

- **Setup:** none
- **Action:** call NormalizeText with `"public"`
- **Expected outcome:** returns `"public"`

---

#### TC-02: Single word

- **Setup:** none
- **Action:** call NormalizeText with `"Interface"`
- **Expected outcome:** returns `"interface"`

---

### Trim

---

#### TC-03: Leading and trailing spaces

- **Setup:** none
- **Action:** call NormalizeText with `"  Interface  "`
- **Expected outcome:** returns `"interface"`

---

#### TC-04: Leading and trailing tabs

- **Setup:** none
- **Action:** call NormalizeText with `"\tInterface\t"`
- **Expected outcome:** returns `"interface"`

---

#### TC-05: Mixed leading whitespace

- **Setup:** none
- **Action:** call NormalizeText with `" \t Interface \t "`
- **Expected outcome:** returns `"interface"`

---

### Collapse

---

#### TC-06: Multiple spaces between words

- **Setup:** none
- **Action:** call NormalizeText with `"Testes   de   aceitacao"`
- **Expected outcome:** returns `"testes de aceitacao"`

---

#### TC-07: Tabs between words

- **Setup:** none
- **Action:** call NormalizeText with `"Testes\tde\taceitacao"`
- **Expected outcome:** returns `"testes de aceitacao"`

---

#### TC-08: Mixed whitespace between words

- **Setup:** none
- **Action:** call NormalizeText with `"Testes \t de \t aceitacao"`
- **Expected outcome:** returns `"testes de aceitacao"`

---

### Case Folding

---

#### TC-09: All uppercase

- **Setup:** none
- **Action:** call NormalizeText with `"PUBLIC"`
- **Expected outcome:** returns `"public"`

---

#### TC-10: Mixed case

- **Setup:** none
- **Action:** call NormalizeText with `"PuBLiC"`
- **Expected outcome:** returns `"public"`

---

#### TC-11: Unicode case folding

- **Setup:** none
- **Action:** call NormalizeText with `"TESTES DE ACEITACAO"`
- **Expected outcome:** returns `"testes de aceitacao"`

---

#### TC-12: German sharp s

- **Setup:** none
- **Action:** call NormalizeText with `"Strasse"`
- **Expected outcome:** returns `"strasse"`
  (Unicode simple case folding maps sharp-s to `"ss"`)

---

### Combined

---

#### TC-13: Trim, collapse, and case fold together

- **Setup:** none
- **Action:** call NormalizeText with `"  TESTES   DE   ACEITACAO  "`
- **Expected outcome:** returns `"testes de aceitacao"`

---

#### TC-14: Logical name qualifier style

- **Setup:** none
- **Action:** call NormalizeText with `"testes de ACEITACAO"`
- **Expected outcome:** returns `"testes de aceitacao"`

---

#### TC-15: Tabs and mixed case

- **Setup:** none
- **Action:** call NormalizeText with `"\tROOT/payments/fees\t"`
- **Expected outcome:** returns `"root/payments/fees"`

---

### Edge Cases

---

#### TC-16: Empty string

- **Setup:** none
- **Action:** call NormalizeText with `""`
- **Expected outcome:** returns `""`

---

#### TC-17: Only whitespace

- **Setup:** none
- **Action:** call NormalizeText with `"   \t  "`
- **Expected outcome:** returns `""`

---

#### TC-18: Non-breaking space is not whitespace

- **Setup:** none
- **Action:** call NormalizeText with a string containing
  `"hello"`, a non-breaking space character (U+00A0), and `"world"`
- **Expected outcome:** returns `"hello world"`
  (the non-breaking space is treated as text, not collapsed or trimmed)

---

#### TC-19: Single character

- **Setup:** none
- **Action:** call NormalizeText with `"X"`
- **Expected outcome:** returns `"x"`
