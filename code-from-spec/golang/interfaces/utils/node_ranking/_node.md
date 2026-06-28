---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
output: code-from-spec/golang/interfaces/utils/node_ranking/output.md
---

# SPEC/golang/interfaces/utils/node_ranking

Iterative ranking of spec tree nodes and artifacts,
with cycle detection as a side effect.

# Public

## Package

`package noderanking`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/noderanking"`

## Interface

```go
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

type NodeRankEntry struct {
	LogicalName string
	Rank        int
}

func NodeRankCompute(entries []*NodeRankInput) ([]*NodeRankEntry, []string, error)
```

Returns ranked entries (nodes and artifacts), cycle
participant logical names, and error.

### Errors

- `ErrUnresolvableReference`: a `depends_on` or `input`
  target cannot be resolved.

# Agent

Generate an interface specification document listing
the package, import path, struct definitions, and
function signatures.
