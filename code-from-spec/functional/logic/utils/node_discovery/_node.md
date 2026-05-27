---
depends_on:
  - ROOT/functional/logic/os/list_files
  - ROOT/functional/logic/utils/logical_names
outputs:
  - id: node_discovery
    path: code-from-spec/functional/logic/utils/node_discovery/output.md
---

# ROOT/functional/logic/utils/node_discovery

Discovers all spec nodes in the spec tree by listing files
and filtering for `_node.md`.

Review status: pending

# Public

## Interface

```
record DiscoveredNode
  logical_name: string
  file_path: CfsPath

function DiscoverNodes() -> list of DiscoveredNode
  errors:
    - directory not found: code-from-spec/ does not exist.
    - walk error: filesystem error while traversing.
    - no nodes found: no _node.md files found.
```

The returned list is sorted alphabetically by logical name.

# Agent

Generate pseudocode for the DiscoverNodes function.

## Implementation guidance

1. Call `ListFiles` with `code-from-spec/` as the directory.
2. Filter the results: keep only files whose name ends
   with `/_node.md`.
3. For each matching file, use `ReverseResolve` from
   `logical_names` to derive the logical name.
4. Sort alphabetically by logical name.
5. If the result is empty, raise "no nodes found".

## Contracts

- Only `_node.md` files are considered nodes.
- The returned list is sorted alphabetically by logical name.
