---
depends_on:
  - ARTIFACT/golang/interfaces/utils/text_normalization
input: ARTIFACT/functional/tests/utils/text_normalization
output: internal/textnormalization/textnormalization_test.go
---

# ROOT/golang/tests/utils/text_normalization

Unit tests for the textnormalization package.

# Agent

## Context

Pure function tests — no filesystem or temp directories needed.
