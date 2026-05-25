---
depends_on:
  - ROOT/functional/utils/logical_names
  - ROOT/functional/utils/frontmatter
outputs:
  - id: cycle_detection
    path: artifacts/functional/utils/cycle_detection/output.md
---

# ROOT/functional/utils/cycle_detection

Detects circular references in the spec tree.

# Public

## Interface

```
function DetectCycles(nodes) -> list of list of string
  errors:
    - unresolvable reference: a depends_on or input target cannot be resolved.
```

Takes the full set of discovered nodes with their parsed
frontmatter. Returns a list of cycles, where each cycle is
a list of logical names forming the circular path
(e.g., `[A, B, C, A]`).

# Agent

## Behavior

### What constitutes a cycle

The spec tree forms a directed graph through four edge
types:

1. **Inheritance** — parent -> child (implicit from tree
   structure).
2. **`depends_on`** — node -> dependency.
3. **`input`** — node -> artifact source node.
4. **`external`** — not a cycle risk (points to files
   outside the spec tree).

A cycle exists when following these edges leads back to
a node already in the path. `external` references are
excluded from cycle detection since they point to project
files, not spec nodes.

### Algorithm

For each node in the tree, perform a depth-first traversal
following `depends_on` and `input` edges. Track the current
path. If a node is encountered that is already in the
current path, a cycle is found.

Inheritance edges (parent -> child) are implicit and do not
need explicit traversal — the v3 spec already prohibits
`depends_on` from pointing to ancestors or descendants.
The cycle detector should verify this rule: a `depends_on`
entry must not point to an ancestor or descendant of the
current node.

## Contracts

- All cycles are reported — not just the first one found.
- Each cycle includes the repeated node at both ends of
  the path list.
