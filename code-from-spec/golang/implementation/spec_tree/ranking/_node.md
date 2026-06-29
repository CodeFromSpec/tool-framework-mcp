---
depends_on:
  - SPEC/golang/implementation/parsing(interface)
output: internal/noderanking/noderanking.go
---

# SPEC/golang/implementation/spec_tree/ranking

Iterative ranking of spec tree nodes and artifacts,
with cycle detection as a side effect.

# Public

## Package

`package noderanking`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/noderanking"`

## Interface

```go
type NodeRankEntry struct {
    Reference parsing.CfsReference
    Rank      int
}

func NodeRankCompute(entries []parsing.Node) ([]NodeRankEntry, []string, error)
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

For each node in entries:
  Add a spec entry keyed by
  node.Reference.LogicalName with:
    - ref: node.Reference
    - deps: empty list (to be filled in step 2)
    - rank: 0

  If node.Frontmatter is not nil and
  node.Frontmatter.Output is not nil:
    Construct artifact logical name: Strip "SPEC/"
    prefix from node.Reference.LogicalName and prepend
    "ARTIFACT/".
    Call parsing.CfsReferenceFromName(artifact logical
    name). If it fails, raise ErrUnresolvableReference.
    Let `artifact_ref` be the result.
    Add an artifact entry keyed by that artifact
    logical name with:
      - ref: *artifact_ref
      - deps: list containing the generating node's
        logical name
      - rank: 0

### Step 2 — Build dependency edges

For each spec node entry in the entry map:
  Find the corresponding node to get its
  Reference.ParentName field.

  If ParentName is nil: Skip — root node has no
  parent dependency.

  Else:
    a. Parent dependency: Add *ParentName to the
       entry's deps list.

    b. depends_on dependencies: If node.Frontmatter is
       not nil, for each reference in
       node.Frontmatter.DependsOn:
         If reference starts with "SPEC/":
           Call parsing.CfsReferenceFromName(reference).
           If it fails, raise ErrUnresolvableReference.
           Let `dep_ref` be the result.
           If dep_ref.LogicalName is not a key in the
           entry map:
             Raise ErrUnresolvableReference
           Add dep_ref.LogicalName to the entry's deps
           list.
         Else if reference starts with "ARTIFACT/":
           If reference is not a key in the entry map:
             Raise ErrUnresolvableReference
           Add reference to the entry's deps list.
         Else if reference starts with "EXTERNAL/":
           Skip — external files have no rank.
         Else:
           Raise ErrUnresolvableReference

    c. input dependency: If node.Frontmatter is not nil
       and node.Frontmatter.Input is not nil:
         If *node.Frontmatter.Input starts with "SPEC/":
           Call parsing.CfsReferenceFromName(
           *node.Frontmatter.Input).
           If it fails, raise ErrUnresolvableReference.
           Let `input_ref` be the result.
           If input_ref.LogicalName is not a key in the
           entry map:
             Raise ErrUnresolvableReference
           Add input_ref.LogicalName to the entry's deps
           list.
         Else if *node.Frontmatter.Input starts with
         "ARTIFACT/":
           If *node.Frontmatter.Input is not a key in
           the entry map:
             Raise ErrUnresolvableReference
           Add *node.Frontmatter.Input to the entry's
           deps list.
         Else if *node.Frontmatter.Input starts with
         "EXTERNAL/": Skip — external files have no
         rank.

### Step 3 — Initialize ranks

All entries start with rank 0 from step 1. Root nodes
(those whose ParentName is nil) keep rank 0 — they
have no parent dependency.

### Step 4 — Iterate and detect cycles

Let N = total number of entries in the entry map.
Let cycle_candidates = empty list.

Repeat up to N times, tracking iteration index i
from 1 to N:
  Let changed = false.

  For each entry in the entry map that has a non-empty
  deps list (root nodes with no deps are skipped):
    Let max_dep_rank = maximum rank among all entries
    in the entry's deps list.
    Let new_rank = 1 + max_dep_rank.

    If new_rank > entry's current rank:
      Update entry's rank to new_rank.
      Set changed = true.
      If i equals N:
        Add entry's logical name to cycle_candidates.

  If changed is false:
    Stop iteration (converged, no cycles).

If iteration completed all N passes and changed was
still true on pass N: Set cycles = cycle_candidates.
Else: Set cycles = empty list.

### Step 5 — Output

Build ranked list: For each entry in the entry map:
  Append NodeRankEntry with Reference = entry.ref,
  Rank = entry.rank.

Sort ranked list: Primary sort: Rank ascending.
Secondary sort: Reference.LogicalName ascending.

Return (ranked: ranked list, cycles: cycles).

## Go-specific guidance

- Use the `parsing` package for `Node`,
  `CfsReference`, `CfsReferenceFromName`.
  Use `strings.HasPrefix` for ARTIFACT/ and EXTERNAL/
  classification.
- The package name should be `noderanking`.
- `NodeRankEntry` is the only exported struct in this
  package.
- Return `([]NodeRankEntry, []string, error)` — ranked
  entries, cycle participant logical names, and error.
