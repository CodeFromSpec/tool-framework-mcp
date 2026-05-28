---
depends_on:
  - ARTIFACT/golang/implementation/internal/textnormalization/code(textnormalization)
input: ARTIFACT/functional/tests/utils/text_normalization(text_normalization_tests)
outputs:
  - id: textnormalization_test
    path: internal/textnormalization/textnormalization_test.go
---

# ROOT/golang/tests/internal/textnormalization

Unit tests for the textnormalization package.

# Agent

## Context

Pure function tests — no filesystem or temp directories needed.
