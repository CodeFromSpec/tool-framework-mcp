---
depends_on:
  - ARTIFACT/golang/interfaces/utils/text_normalization
output: internal/textnormalization/textnormalization_test.go
---

# SPEC/golang/tests/utils/text_normalization

Unit tests for the textnormalization package.

# Agent

## Context

Pure function tests — no filesystem or temp directories
needed.

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

Input: "Testes   de   aceitacao".
Expect: "testes de aceitacao".

#### Tabs between words

Input: "Testes\tde\taceitacao".
Expect: "testes de aceitacao".

#### Mixed whitespace between words

Input: "Testes \t de \t aceitacao".
Expect: "testes de aceitacao".

### Case folding

#### All uppercase

Input: "PUBLIC". Expect: "public".

#### Mixed case

Input: "PuBLiC". Expect: "public".

#### Unicode case folding

Input: "TESTES DE ACEITACAO".
Expect: "testes de aceitacao".

#### German sharp s

Input: "Strasse". Expect: "strasse".

### Combined

#### Trim, collapse, and case fold together

Input: "  TESTES   DE   ACEITACAO  ".
Expect: "testes de aceitacao".

#### Logical name qualifier style

Input: "testes de ACEITACAO".
Expect: "testes de aceitacao".

#### Tabs and mixed case

Input: "\tROOT/payments/fees\t".
Expect: "root/payments/fees".

### Edge cases

#### Empty string

Input: "". Expect: "".

#### Only whitespace

Input: "   \t  ". Expect: "".

#### Non-breaking space is not whitespace

Setup: input = "hello" + U+00A0 + "world" (no regular
spaces).

Expected: "hello world" — the non-breaking space
is preserved as text, case-folded but not collapsed.

#### Single character

Input: "X". Expect: "x".

## Go-specific guidance

- The package name is `textnormalization_test` (external
  test package).
- Pure function tests — no file I/O needed.
