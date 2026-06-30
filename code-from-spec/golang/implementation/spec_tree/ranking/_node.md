---
depends_on:
  - SPEC/golang/implementation/parsing(interface)
output: internal/noderanking/noderanking.go
---

# SPEC/golang/implementation/spec_tree/ranking

Iterative ranking of spec tree nodes and artifacts, with cycle detection
as a side effect.

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

Returns ranked entries (nodes and artifacts), cycle participant logical
names, and error.

### Errors

- `ErrUnresolvableReference`: a `depends_on` or `input` target cannot
  be resolved.

# Agent

Implement the node ranking component as a Go package.

## Logic

### Step 1 — Build entry map

From the input list, build an entry map keyed by logical name. Each
entry tracks its dependency list and current rank.

For each node in entries:

1. Add a spec entry keyed by node.Reference.LogicalName with:
   - ref: node.Reference
   - deps: empty list (to be filled in step 2)
   - rank: 0

2. If node.Frontmatter is not nil and node.Frontmatter.Output is not
   nil:
   - Construct artifact logical name: strip "SPEC/" prefix from
     node.Reference.LogicalName and prepend "ARTIFACT/".
   - Construct a CfsReference directly:
     - NodeType: parsing.CfsNodeTypeArtifact
     - LogicalName: the artifact logical name
     - Qualifier: nil
     - Path: *node.Frontmatter.Output
     - ParentName: pointer to node.Reference.LogicalName
   - Add an artifact entry keyed by that artifact logical name with:
     - ref: the constructed CfsReference
     - deps: list containing the generating node's logical name
     - rank: 0

### Step 2 — Build dependency edges

For each spec node entry in the entry map:

1. Find the corresponding node to get its Reference.ParentName and
   Frontmatter fields.

2. **Parent dependency**: If ParentName is not nil, add *ParentName to
   the entry's deps list.

3. **depends_on dependencies**: If node.Frontmatter is not nil, for
   each reference in node.Frontmatter.DependsOn:
   - If reference starts with "SPEC/":
     - Extract the unqualified logical name: if the reference contains
       "(", take the portion before it; otherwise use the reference
       as-is.
     - If the unqualified name is not a key in the entry map, raise
       ErrUnresolvableReference.
     - Add the unqualified name to the entry's deps list.
   - Else if reference starts with "ARTIFACT/":
     - If reference is not a key in the entry map, raise
       ErrUnresolvableReference.
     - Add reference to the entry's deps list.
   - Else if reference starts with "EXTERNAL/": skip.
   - Else: raise ErrUnresolvableReference.

4. **input dependency**: If node.Frontmatter is not nil and
   node.Frontmatter.Input is not nil:
   - If *node.Frontmatter.Input starts with "SPEC/":
     - Extract the unqualified logical name: if the value contains "(",
       take the portion before it; otherwise use as-is.
     - If the unqualified name is not a key in the entry map, raise
       ErrUnresolvableReference.
     - Add the unqualified name to the entry's deps list.
   - Else if starts with "ARTIFACT/":
     - If not a key in the entry map, raise ErrUnresolvableReference.
     - Add to the entry's deps list.
   - Else if starts with "EXTERNAL/": skip.

### Step 3 — Initialize ranks

All entries start with rank 0.

### Step 4 — Iterate and detect cycles

- Let N = total number of entries in the entry map.
- Let cycle_candidates = empty list.
- Repeat up to N times, tracking iteration index i from 1 to N:
  - Let changed = false.
  - For each entry in the entry map that has a non-empty deps list:
    - Let max_dep_rank = maximum rank among all entries in the entry's
      deps list.
    - Let new_rank = 1 + max_dep_rank.
    - If new_rank > entry's current rank:
      - Update entry's rank to new_rank.
      - Set changed = true.
      - If i equals N: add entry's logical name to cycle_candidates.
  - If changed is false: stop iteration (converged, no cycles).
- If iteration completed all N passes and changed was still true on
  pass N: set cycles = cycle_candidates.
- Else: set cycles = empty list.

### Step 5 — Output

- Build ranked list: for each entry in the entry map, append
  NodeRankEntry with Reference = entry.ref, Rank = entry.rank.
- Sort ranked list: primary sort by Rank ascending, secondary sort by
  Reference.LogicalName ascending.
- Return (ranked list, cycles, nil).

## Go-specific guidance

- Use the `parsing` package for `Node`, `CfsReference`,
  `CfsNodeTypeArtifact`. Do not call `CfsReferenceFromName` — construct
  CfsReference values directly. Use `strings.HasPrefix` for SPEC/,
  ARTIFACT/, and EXTERNAL/ classification. Use `strings.Index` to find
  "(" for qualifier extraction.
- The package name should be `noderanking`.
- `NodeRankEntry` is the only exported struct in this package.
- Return `([]NodeRankEntry, []string, error)` — ranked entries, cycle
  participant logical names, and error.
