---
depends_on:
  - ROOT/golang/dependencies/golang-x-text
input: ARTIFACT/functional/logic/utils/name_normalization(name_normalization)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: normalizename
    path: internal/normalizename/normalizename.go
---

# ROOT/golang/implementation/internal/normalizename/code

Generates the normalizename package implementation.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `golang.org/x/text` for Unicode case folding and
  normalization as described in the input.
- The package name should be `normalizename`.
