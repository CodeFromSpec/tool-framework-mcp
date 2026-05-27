---
depends_on:
  - ROOT/golang/implementation/internal/logical_names
  - ROOT/golang/implementation/internal/frontmatter
input: ARTIFACT/functional/logic/utils/node_ranking(node_ranking)
outputs:
  - id: noderanking
    path: internal/noderanking/noderanking.go
---

# ROOT/golang/implementation/internal/node_ranking/code

Generates the noderanking package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Depends on the `logicalnames`, `frontmatter`, and
  `nodediscovery` packages.
- `DetectCycles` receives `[]nodediscovery.DiscoveredNode`
  which has only `LogicalName` and `FilePath` — it does
  not carry parsed frontmatter. Parse frontmatter
  internally using `frontmatter.ParseFrontmatter(node.FilePath)`
  for each node before building the entry map.
- Use the iterative ranking algorithm described in the
  functional spec: initialize all ranks to 0, iterate
  until convergence, detect cycles by checking for rank
  changes after N passes.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
