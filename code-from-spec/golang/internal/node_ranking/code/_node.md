---
depends_on:
  - ROOT/golang/internal/logical_names
  - ROOT/golang/internal/frontmatter
input: ARTIFACT/functional/utils/node_ranking(node_ranking)
outputs:
  - id: noderanking
    path: internal/noderanking/noderanking.go
---

# ROOT/golang/internal/node_ranking/code

Generates the noderanking package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Depends on the `logicalnames` and `frontmatter` packages.
- Use the iterative ranking algorithm described in the
  functional spec: initialize all ranks to 0, iterate
  until convergence, detect cycles by checking for rank
  changes after N passes.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
