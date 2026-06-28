---
depends_on:
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
output: internal/noderanking/noderanking.go
---

# SPEC/golang/implementation/utils/node_ranking

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

Implement the node ranking component as a Go package.

## Logic

### Step 1 — Build entry map

From the input list, build an entry map keyed by logical
name. Each entry tracks its dependency list and current
rank.

For each NodeRankInput in entries:
  Add a spec entry keyed by logical_name with:
    - deps: empty list (to be filled in step 2)
    - rank: 0

  If frontmatter.output is non-empty:
    Construct artifact logical name: Strip "SPEC/"
    prefix from logical_name and prepend "ARTIFACT/".
    Add an artifact entry keyed by that artifact
    logical name with:
      - deps: list containing the generating node's
        logical_name
      - rank: 0

### Step 2 — Build dependency edges

For each spec node entry in the entry map:
  If logical_name is "SPEC": Skip — root node has no
  dependencies.

  Else:
    a. Parent dependency: Call
       LogicalNameParse(logical_name). Let `ln` be
       the result. If it fails, raise error
       "UnresolvableReference". Add *ln.Parent to
       the entry's deps list.

    b. depends_on dependencies: For each reference in
       frontmatter.depends_on:
         If reference starts with "SPEC/" or equals
         "SPEC":
           Call LogicalNameParse(reference). If it
           fails, raise error "UnresolvableReference".
           Let `dep_ln` be the result.
           If dep_ln.Name is not a key in the entry
           map:
             Raise error "UnresolvableReference"
           Add dep_ln.Name to the entry's deps list.
         Else if reference starts with "ARTIFACT/":
           If reference is not a key in the entry map:
             Raise error "UnresolvableReference"
           Add reference to the entry's deps list.
         Else if reference starts with "EXTERNAL/":
           Skip — external files have no rank.
         Else:
           Raise error "UnresolvableReference"

    c. input dependency: If frontmatter.input is
       non-empty:
         If frontmatter.input starts with "ARTIFACT/":
           If frontmatter.input is not a key in the
           entry map:
             Raise error "UnresolvableReference"
           Add frontmatter.input to the entry's deps
           list.
         Else if frontmatter.input starts with
         "EXTERNAL/": Skip — external files have no
         rank.

### Step 3 — Initialize ranks

Set rank of entry keyed "SPEC" to 0 (fixed, never
updated). All other entries already have rank 0 from
step 1.

### Step 4 — Iterate and detect cycles

Let N = total number of entries in the entry map.
Let cycle_candidates = empty list.

Repeat up to N times, tracking iteration index i
from 1 to N:
  Let changed = false.

  For each entry in the entry map (excluding "SPEC"):
    Let max_dep_rank = maximum rank among all entries
    in the entry's deps list.
    Let new_rank = 1 + max_dep_rank.

    If new_rank > entry's current rank:
      Update entry's rank to new_rank.
      Set changed = true.
      If i equals N:
        Add entry's logical_name to cycle_candidates.

  If changed is false:
    Stop iteration (converged, no cycles).

If iteration completed all N passes and changed was
still true on pass N: Set cycles = cycle_candidates.
Else: Set cycles = empty list.

### Step 5 — Output

Build ranked list: For each entry in the entry map:
  Append NodeRankEntry with logical_name and rank.

Sort ranked list: Primary sort: rank ascending.
Secondary sort: logical_name ascending.

Return (ranked: ranked list, cycles: cycles).

## Go-specific guidance

- Use the `frontmatter` package for the `Frontmatter`
  record.
- Use the `logicalnames` package for
  `LogicalNameParse`, `NodeTypeSpec`,
  `NodeTypeArtifact`, `NodeTypeExternal`.
- The package name should be `noderanking`.
- `NodeRankInput` and `NodeRankEntry` are exported
  structs in this package.
- Return `([]NodeRankEntry, []string, error)` — ranked
  entries, cycle participant logical names, and error.
