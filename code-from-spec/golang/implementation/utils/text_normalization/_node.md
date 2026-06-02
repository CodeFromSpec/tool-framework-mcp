---
depends_on:
  - ROOT/golang/dependencies/golang-x-text
  - ARTIFACT/golang/interfaces/utils/text_normalization
input: ARTIFACT/functional/logic/utils/text_normalization
output: internal/textnormalization/textnormalization.go
---

# ROOT/golang/implementation/utils/text_normalization

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `golang.org/x/text` for Unicode case folding and
  normalization as described in the input.
- The package name should be `textnormalization`.
