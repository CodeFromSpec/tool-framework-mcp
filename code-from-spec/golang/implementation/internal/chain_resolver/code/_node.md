---
depends_on:
  - ROOT/golang/implementation/internal/frontmatter
  - ROOT/golang/implementation/internal/logical_names
input: ARTIFACT/functional/logic/utils/chain_resolver(chain_resolver)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: chainresolver
    path: internal/chainresolver/chainresolver.go
---

# ROOT/golang/implementation/internal/chain_resolver/code

Generates the chainresolver package implementation.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `logicalnames` and `frontmatter` packages for
  resolution and parsing.
- Use `os.Stat` to verify files exist on disk.
- Use `filepath.ToSlash` for path normalization.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
