<!-- code-from-spec: ROOT/functional/logic/spec_tree/scan@BOCN_cdbsKfp9njAI5EFiFsp4Q4 -->

# Spec Tree Scan

namespace: spectreescan

## Records

record SpecTreeNode
  logical_name: string
  file_path: pathutils.PathCfs

## Functions

```
function SpecTreeScan() -> list of SpecTreeNode
  errors:
    - NoNodesFound: no _node.md files were found under
      code-from-spec/.
    - (ListFiles.*): propagated from ListFiles.
    - (LogicalNames.*): propagated from LogicalNameFromPath.

  1. Call ListFiles with "code-from-spec/" as the directory.
     If ListFiles raises an error, propagate it unchanged.

  2. Filter the returned file paths: keep only those where
     the file name portion (everything after the last "/")
     is exactly "_node.md".

  3. For each remaining path in the filtered list:
     a. Call LogicalNameFromPath with the path.
        If LogicalNameFromPath raises an error, propagate it
        unchanged.
     b. Construct a SpecTreeNode with:
          logical_name: the result of LogicalNameFromPath
          file_path: the current path

  4. Sort the list of SpecTreeNode records alphabetically
     by logical_name.

  5. If the sorted list is empty, raise error "no nodes found".

  6. Return the sorted list.
```
