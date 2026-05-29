<!-- code-from-spec: ROOT/functional/logic/spec_tree/scan@plEwz9NNJyrpO8W8djf-_bUClII -->

## Data Structures

record SpecTreeNode
  logical_name: string
  file_path: PathCfs

## Functions

function SpecTreeScan() -> list of SpecTreeNode

  1. Call ListFiles with "code-from-spec/" as the directory.
     If ListFiles raises an error, propagate it.

  2. Filter the results: keep only entries where the file name
     is exactly "_node.md".
     The file name is the portion of the path after the last "/".

  3. For each matching file path:
     a. Call LogicalNameFromPath with the file path.
        If LogicalNameFromPath raises an error, propagate it.
     b. Create a SpecTreeNode record with:
        - logical_name: the result from LogicalNameFromPath
        - file_path: the matching file path

  4. Sort the list of SpecTreeNode records alphabetically by
     logical_name.

  5. If the sorted list is empty, raise error "no nodes found".

  6. Return the sorted list.

## Error Conditions

- If ListFiles raises any error, propagate it unchanged.
- If LogicalNameFromPath raises any error for any file, propagate
  it unchanged.
- If no _node.md files are found under "code-from-spec/", raise
  error "no nodes found".
