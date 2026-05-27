---
depends_on:
  - ROOT/golang/implementation/internal/logical_names
  - ROOT/golang/implementation/internal/frontmatter
  - ROOT/golang/implementation/internal/parsenode
  - ROOT/golang/implementation/internal/pathvalidation
  - ARTIFACT/golang/interfaces/internal/file_reader(interface)
  - ROOT/golang/implementation/internal/normalizename
input: ARTIFACT/functional/logic/utils/format_validation(format_validation)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: formatvalidation
    path: internal/formatvalidation/formatvalidation.go
---

# ROOT/golang/implementation/internal/format_validation/code

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
