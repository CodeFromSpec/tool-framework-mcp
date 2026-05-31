---
depends_on:
  - ROOT/functional/logic/os/list_files(interface)
  - ROOT/functional/logic/os/path_utils(interface)
  - ROOT/functional/logic/utils/logical_names(interface)
outputs:
  - id: spec_tree
    path: code-from-spec/functional/logic/spec_tree/scan/output.md
---

# ROOT/functional/logic/spec_tree/scan

Scans the `code-from-spec/` directory and returns all
spec nodes found.

# Public

## Interface

```
record SpecTreeNode
  logical_name: string
  file_path: PathCfs

function SpecTreeScan() -> list of SpecTreeNode
  errors:
    - NoNodesFound: no _node.md files found under
      code-from-spec/.
    - (ListFiles.*): propagated from ListFiles.
    - (LogicalNames.*): propagated from
      LogicalNameFromPath.
```

`SpecTreeScan` takes no parameters. It always scans the
`code-from-spec/` directory relative to the project root.

The returned list is sorted alphabetically by logical name.

# Agent

Generate pseudocode for the SpecTreeScan function.

## Implementation guidance

1. Call `ListFiles` with `code-from-spec/` as the
   directory. If `ListFiles` raises an error, propagate it.
2. Filter the results: keep only files whose file name
   is exactly `_node.md`. The file name is everything
   after the last `/` in the path.
3. For each matching file, call `LogicalNameFromPath` to
   derive the logical name. If it raises an error,
   propagate it.
4. Sort the result alphabetically by logical name.
5. If the result is empty, raise "no nodes found".

## Contracts

- Only files named exactly `_node.md` are considered nodes.
- The returned list is sorted alphabetically by logical name.
- Only scans the `code-from-spec/` directory.
