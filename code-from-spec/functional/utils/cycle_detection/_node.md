---
outputs:
  - id: cycle_detection
    path: code-from-spec/functional/utils/cycle_detection/output.md
---

# ROOT/functional/utils/cycle_detection

Detects circular references in the spec tree.

# Public

## Behavior

### Input

The full set of discovered nodes with their parsed
frontmatter.

### Output

A list of cycles. Each cycle is a list of logical names
forming the circular path (e.g., `[A, B, C, A]`).

## What constitutes a cycle

The spec tree forms a directed graph through four edge
types:

1. **Inheritance** — parent → child (implicit from tree
   structure).
2. **`depends_on`** — node → dependency.
3. **`input`** — node → artifact source node.
4. **`external`** — not a cycle risk (points to files
   outside the spec tree).

A cycle exists when following these edges leads back to
a node already in the path. `external` references are
excluded from cycle detection since they point to project
files, not spec nodes.

## Algorithm

For each node in the tree, perform a depth-first traversal
following `depends_on` and `input` edges. Track the current
path. If a node is encountered that is already in the
current path, a cycle is found.

Inheritance edges (parent → child) are implicit and do not
need explicit traversal — the v3 spec already prohibits
`depends_on` from pointing to ancestors or descendants.
The cycle detector should verify this rule: a `depends_on`
entry must not point to an ancestor or descendant of the
current node.

## Error conditions

| Condition | Description |
|---|---|
| Unresolvable reference | A `depends_on` or `input` target cannot be resolved. |
