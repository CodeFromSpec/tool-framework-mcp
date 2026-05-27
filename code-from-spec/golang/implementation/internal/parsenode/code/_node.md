---
depends_on:
  - ROOT/golang/dependencies/yuin-goldmark
  - ROOT/golang/implementation/internal/logical_names
  - ROOT/golang/implementation/internal/normalizename
input: ARTIFACT/functional/logic/utils/node_parsing(node_parsing)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: parsenode
    path: internal/parsenode/parsenode.go
---

# ROOT/golang/implementation/internal/parsenode/code

Generates the parsenode package implementation.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `yuin/goldmark` for CommonMark parsing. See
  `ROOT/golang/dependencies/yuin-goldmark` for heading text
  extraction and raw source extraction helpers.
- Use `normalizename.NormalizeName` for all heading comparisons.
- Use `logicalnames.PathFromLogicalName` to resolve logical
  names to file paths.
- Use `os.ReadFile` to read the file content.
- The package name should be `parsenode`.
