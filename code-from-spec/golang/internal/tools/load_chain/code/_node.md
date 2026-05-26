---
depends_on:
  - ROOT/golang/dependencies/google-uuid
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/internal/chain_resolver
  - ROOT/golang/internal/frontmatter
  - ROOT/golang/internal/logical_names
  - ROOT/golang/internal/normalizename
  - ROOT/golang/internal/parsenode
  - ROOT/golang/internal/pathvalidation
input: ARTIFACT/functional/mcp_tools/load_chain(load_chain)
outputs:
  - id: load_chain
    path: internal/load_chain/load_chain.go
---

# ROOT/golang/internal/tools/load_chain/code

Implementation of the load_chain tool handler.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `github.com/google/uuid` for UUID generation.
- Use the `mcp-go` SDK types for tool results.
- Call internal packages (`chainresolver`, `frontmatter`,
  `logicalnames`, `normalizename`, `parsenode`,
  `pathvalidation`) for their respective operations.
- Use `os.ReadFile` to read chain files.
- The package name should be `load_chain`.

## Constraints

- The target argument must be a logical name that resolves to a
  node with `outputs`. Absent, empty, or invalid values cause
  the tool to report an error.
- Reads are limited to the chain.
- If any chain file cannot be read, `load_chain` returns an error
  identifying the missing file; it does not return partial results.
