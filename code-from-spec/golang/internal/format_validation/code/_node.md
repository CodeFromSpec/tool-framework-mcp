---
depends_on:
  - ROOT/golang/internal/logical_names
  - ROOT/golang/internal/frontmatter
  - ROOT/golang/internal/parsenode
  - ROOT/golang/internal/pathvalidation
  - ARTIFACT/golang/internal/file_reader/interface(interface)
  - ROOT/golang/internal/normalizename
input: ARTIFACT/functional/utils/format_validation(format_validation)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: formatvalidation
    path: internal/formatvalidation/formatvalidation.go
---

# ROOT/golang/internal/format_validation/code

Generates the formatvalidation package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Depends on: `logicalnames`, `frontmatter`, `parsenode`,
  `pathvalidation`, `filereader`, `normalizename` packages.
- Returns `[]FormatError` slice collecting all violations.
  Do not stop at the first error.
- A node has children if any other discovered node's logical
  name starts with its logical name followed by `/`.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
