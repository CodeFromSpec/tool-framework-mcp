<!-- code-from-spec: ROOT/functional/logic/utils/spec_tree@plEwz9NNJyrpO8W8djf-_bUClII -->

## Records

record SpecTreeNode
  logical_name: string
  file_path: PathCfs


## Functions

function SpecTreeScan() -> list of SpecTreeNode
  errors:
    - (list errors): propagated from ListFiles.
    - (name errors): propagated from LogicalNameFromPath.
    - no nodes found: no _node.md files found under code-from-spec/.

  1. Call ListFiles with the path "code-from-spec/".
     If ListFiles raises an error, propagate it.

  2. Filter the resulting list.
     Keep only entries where the file name portion equals "_node.md".
     The file name is the substring after the last "/" in the path value.

  3. For each remaining PathCfs in the filtered list:
     a. Call LogicalNameFromPath with the PathCfs.
        If LogicalNameFromPath raises an error, propagate it.
     b. Create a SpecTreeNode record:
          logical_name: the string returned by LogicalNameFromPath
          file_path: the current PathCfs

  4. Sort the list of SpecTreeNode records alphabetically by logical_name.

  5. If the sorted list is empty, raise error "no nodes found".

  6. Return the sorted list.
