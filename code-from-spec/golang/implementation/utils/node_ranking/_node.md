---
depends_on:
  - ARTIFACT/golang/interfaces/utils/node_ranking
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/logic/utils/node_ranking
output: internal/noderanking/noderanking.go
---

# ROOT/golang/implementation/utils/node_ranking

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `frontmatter` package for the `Frontmatter` record.
- Use the `logicalnames` package for `LogicalNameGetParent`.
- The package name should be `noderanking`.
- `NodeRankInput` and `NodeRankEntry` are exported structs
  in this package.
- Return `([]NodeRankEntry, []string, error)` — ranked
  entries, cycle participant logical names, and error.
