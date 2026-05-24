---
outputs:
  - id: tree_walker
    path: code-from-spec/functional/utils/tree_walker/output.md
---

# ROOT/functional/utils/tree_walker

Walks the filesystem to discover all spec nodes in the
spec tree.

# Public

## Behavior

### Input

No parameters. Starts from `code-from-spec/` relative to
the project root (working directory).

### Output

A list of discovered nodes, each with:
- `logical_name` — derived from the filesystem path.
- `file_path` — path relative to project root.

The list is sorted alphabetically by logical name.

## Discovery rules

Walk `code-from-spec/` recursively. Every `_node.md` file
produces a discovered node. Other files are ignored.

For each `_node.md` found, use reverse resolution (see
`ROOT/functional/utils/logical_names`) to derive the logical
name from the file path.

## Node classification

After discovery, each node can be classified by checking
whether it has child directories containing `_node.md`
files:
- **Leaf node** — no children with `_node.md`.
- **Intermediate node** — has children with `_node.md`.
- **Root node** — the `ROOT` node itself.

## Error conditions

| Condition | Description |
|---|---|
| Directory not found | `code-from-spec/` does not exist. |
| Walk error | Filesystem error while traversing. |
| No nodes found | `code-from-spec/` contains no `_node.md` files. |
