<!-- code-from-spec: ROOT/functional/logic/spec_tree/scan@7bPGmqw0RGyeO74_rMRJMflBYh4 -->

namespace: spectreescan

record SpecTreeNode
  logical_name: string
  file_path: pathutils.PathCfs

function SpecTreeScan() -> list of SpecTreeNode
  errors:
    - NoNodesFound: no _node.md files were found under code-from-spec/.
    - (ListFiles.*): propagated from ListFiles.
    - (LogicalNames.*): propagated from LogicalNameFromPath.

  1. Call ListFiles with PathCfs value "code-from-spec/".
     If ListFiles raises an error, propagate it.

  2. Filter the resulting list: keep only files whose file name is exactly "_node.md".
     The file name is the portion of the path after the last "/".

  3. For each file that passed the filter, determine whether it resides inside
     a "_"-prefixed directory directly under "code-from-spec/":
       a. Remove the leading "code-from-spec/" prefix from the file's path value.
       b. Find the first "/" in the remaining string.
          If there is no "/", the file is directly inside "code-from-spec/" —
          do not exclude it.
          If there is a "/", the text before it is the first directory segment.
          If that segment starts with "_", exclude the file.
          Otherwise, keep the file.

  4. For each file that was not excluded, call LogicalNameFromPath with the file's PathCfs.
     If LogicalNameFromPath raises an error, propagate it.
     Construct a SpecTreeNode record with:
       - logical_name: the string returned by LogicalNameFromPath
       - file_path: the PathCfs of the file

  5. Sort the collected SpecTreeNode records alphabetically by logical_name.

  6. If the sorted list is empty, raise error "no nodes found".

  7. Return the sorted list of SpecTreeNode records.
