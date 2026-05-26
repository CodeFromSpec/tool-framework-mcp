---
input: ARTIFACT/functional/utils/logical_names(logical_names)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: logicalnames
    path: internal/logicalnames/logicalnames.go
---

# ROOT/golang/internal/logical_names/code

Generates the logicalnames package implementation.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `filepath` and `path` standard library packages for
  path manipulation.
- The package name should be `logicalnames`.
