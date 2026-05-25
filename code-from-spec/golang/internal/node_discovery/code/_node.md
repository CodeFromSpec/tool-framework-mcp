---
depends_on:
  - ROOT/golang/internal/logical_names
input: ARTIFACT/functional/utils/node_discovery(node_discovery)
outputs:
  - id: nodediscovery
    path: internal/nodediscovery/nodediscovery.go
---

# ROOT/golang/internal/node_discovery/code

Generates the nodediscovery package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `filepath.WalkDir` for filesystem traversal.
- Depends on the `logicalnames` package for reverse
  resolution from file paths to logical names.
- Sort the result slice alphabetically by `LogicalName`.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
