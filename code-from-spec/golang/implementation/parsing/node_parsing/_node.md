---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/utils/text_normalization
input: ARTIFACT/functional/logic/parsing/node_parsing
output: internal/parsenode/parsenode.go
---

# SPEC/golang/implementation/parsing/node_parsing

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `textnormalization.NormalizeText` for all heading
  comparisons.
- Use `logicalnames.LogicalNameToPath` to resolve logical
  names to file paths.
- Use the `file` package for file I/O: `FileOpen`,
  `FileReadLine`, `FileClose`.
- The package name should be `parsenode`.
