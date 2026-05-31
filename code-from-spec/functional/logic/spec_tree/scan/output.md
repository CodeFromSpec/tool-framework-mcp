<!-- code-from-spec: ROOT/functional/logic/spec_tree/scan@I4pQwkS97PBjpp8zVoD8XLoJobw -->

## Records

```
record SpecTreeNode
  logical_name: string
  file_path: PathCfs
```

## Functions

```
function SpecTreeScan() -> list of SpecTreeNode
  errors:
    - NoNodesFound: no _node.md files were found under
      code-from-spec/.
    - (ListFiles.*): propagated from ListFiles.
    - (LogicalNames.*): propagated from LogicalNameFromPath.
```

Scans the `code-from-spec/` directory and returns all spec
nodes found, sorted alphabetically by logical name.

### Steps

1. Call `ListFiles` with `"code-from-spec/"` as the directory.
   If `ListFiles` raises an error, propagate it unchanged.

2. Filter the returned list.
   For each file path in the list:
     Extract the file name — everything after the last `"/"` in the path value.
     Keep the file only if the file name is exactly `"_node.md"`.
   The result is the filtered list of node file paths.

3. For each file path in the filtered list:
     Call `LogicalNameFromPath` with the file path.
     If `LogicalNameFromPath` raises an error, propagate it unchanged.
     Create a `SpecTreeNode` record with:
       logical_name: the value returned by `LogicalNameFromPath`
       file_path: the current file path
   The result is the list of `SpecTreeNode` records.

4. Sort the list of `SpecTreeNode` records alphabetically by
   the `logical_name` field.

5. If the sorted list is empty, raise error `"no nodes found"`.

6. Return the sorted list of `SpecTreeNode` records.
```
