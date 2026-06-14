---
depends_on:
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/utils/logical_names
output: internal/logicalnames/logicalnames.go
---

# SPEC/golang/implementation/utils/logical_names

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `filepath` and `path` standard library packages for
  path manipulation.
- The package name should be `logicalnames`.
- Functions that declare errors in the functional spec
  should return `(result, error)` in Go.
- Functions that return `optional` in the functional spec
  should return `(result, bool)` in Go.
- Boolean functions return a single `bool`.
