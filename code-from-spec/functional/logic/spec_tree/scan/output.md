<!-- code-from-spec: ROOT/functional/logic/spec_tree/scan@o5zywBGePnW68eCf3RxajBMkVXg -->

## Namespace

    namespace: spectreescan

## Records

```
record SpecTreeNode
  logical_name: string
  file_path: pathutils.PathCfs
```

## Functions

```
function SpecTreeScan() -> list of SpecTreeNode
  errors:
    - NoNodesFound: no _node.md files found under code-from-spec/.
    - (ListFiles.*): propagated from ListFiles.
    - (LogicalNames.*): propagated from LogicalNameFromPath.
```

### SpecTreeScan

  1. Call `ListFiles` with the path `"code-from-spec/"`.
     If `ListFiles` raises an error, propagate it.

  2. Filter the results: keep only entries where the file name portion
     of the path (everything after the last `"/"`) is exactly `"_node.md"`.

  3. For each matching file path, call `LogicalNameFromPath` to derive
     the logical name.
     If `LogicalNameFromPath` raises an error, propagate it.
     Build a `SpecTreeNode` record with:
       `logical_name`: the derived logical name
       `file_path`: the matching file path

  4. Sort the resulting list alphabetically by `logical_name`.

  5. If the list is empty, raise error `NoNodesFound`.

  6. Return the sorted list of `SpecTreeNode` records.
