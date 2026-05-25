---
depends_on:
  - ROOT/functional/utils/logical_names
outputs:
  - id: node_discovery
    path: code-from-spec/functional/utils/node_discovery/output.md
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

function DiscoverNodes() -> list of DiscoveredNode
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

Find every `_node.md` file inside `code-from-spec/` and
all its subdirectories. Each `_node.md` file produces a
discovered node. Other files are ignored.

For each `_node.md` found, use reverse resolution (see
`ROOT/functional/utils/logical_names`) to derive the logical
name from the file path.

## Contracts

- The returned list is sorted alphabetically by logical name.
- Only `_node.md` files are considered nodes.
