<!-- code-from-spec: ROOT/functional/logic/spec_tree/scan@plEwz9NNJyrpO8W8djf-_bUClII -->

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

  2. Filter the returned list: keep only entries where the
     file name is exactly "_node.md".
     The file name is the portion of the path after the last "/".

  3. For each remaining file path in the filtered list:
     a. Call LogicalNameFromPath with the file path.
        If LogicalNameFromPath raises an error, propagate it.
     b. Create a SpecTreeNode record with:
          logical_name: the result of LogicalNameFromPath
          file_path: the current file path

  4. Sort the resulting list of SpecTreeNode records
     alphabetically by logical_name.

  5. If the sorted list is empty, raise error "no nodes found".

  6. Return the sorted list.
