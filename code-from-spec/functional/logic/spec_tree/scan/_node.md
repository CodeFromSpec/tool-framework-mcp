---
depends_on:
  - ROOT/functional/logic/os/list_files(interface)
  - ROOT/functional/logic/os/path_utils(interface)
  - ROOT/functional/logic/utils/logical_names(interface)
output: code-from-spec/functional/logic/spec_tree/scan/output.md
---

# ROOT/functional/logic/spec_tree/scan

Scans the `code-from-spec/` directory and returns all
spec nodes found.

# Public

## Namespace

    namespace: spectreescan

## Interface

```
record SpecTreeNode
  logical_name: string
  file_path: pathutils.PathCfs

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
3. Exclude files inside `_`-prefixed directories directly
   under `code-from-spec/`. A file is excluded if its
   path, after removing the `code-from-spec/` prefix,
   starts with `_` (e.g. `code-from-spec/_rules/x/_node.md`
   is excluded, but `code-from-spec/a/_b/_node.md` is not —
   only the first path segment after `code-from-spec/`
   is checked).
4. For each remaining file, call `LogicalNameFromPath` to
   derive the logical name. If it raises an error,
   propagate it.
5. Sort the result alphabetically by logical name.
6. If the result is empty, raise "no nodes found".

## Contracts

- Only files named exactly `_node.md` are considered nodes.
- Files inside `_`-prefixed directories directly under
  `code-from-spec/` are ignored (not nodes, not errors).
- The returned list is sorted alphabetically by logical name.
- Only scans the `code-from-spec/` directory.
