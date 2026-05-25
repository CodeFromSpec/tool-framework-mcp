---
depends_on:
  - ROOT/functional/utils/logical_names
outputs:
  - id: node_discovery
    path: artifacts/functional/utils/node_discovery/output.md
---

# ROOT/functional/utils/node_discovery

Walks the filesystem to discover all spec nodes in the
spec tree.

# Public

## Interface

```
record DiscoveredNode
  logical_name: string
  file_path: string

function WalkTree() -> list of DiscoveredNode
  errors:
    - directory not found: code-from-spec/ does not exist.
    - walk error: filesystem error while traversing.
    - no nodes found: code-from-spec/ contains no _node.md files.
```

The returned list is sorted alphabetically by logical name.

# Agent

## Behavior

Starts from `code-from-spec/` relative to the project root
(working directory). No parameters.

### Discovery rules

Walk `code-from-spec/` recursively. Every `_node.md` file
produces a discovered node. Other files are ignored.

For each `_node.md` found, use reverse resolution (see
`ROOT/functional/utils/logical_names`) to derive the logical
name from the file path.

### Node classification

After discovery, each node can be classified by checking
whether it has child directories containing `_node.md`
files:
- **Leaf node** — no children with `_node.md`.
- **Intermediate node** — has children with `_node.md`.
- **Root node** — the `ROOT` node itself.

## Contracts

- The returned list is sorted alphabetically by logical name.
- Only `_node.md` files are considered nodes.
