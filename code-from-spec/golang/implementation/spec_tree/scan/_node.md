---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/scan
  - ARTIFACT/golang/interfaces/os/list_files
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/utils/logical_names
output: internal/spectree/spectree.go
---

# SPEC/golang/implementation/spec_tree/scan

# Agent

Implement the spec tree scan as a Go package. The output
file is the sole .go file in the package — declare all
types, error sentinels, and function signatures from the
interface artifact in this file.

## Logic

1. Call `ListFiles` with "code-from-spec/" as the
   directory. If `ListFiles` raises an error, propagate
   it.

2. Filter the list: keep only files whose name after
   the last "/" is exactly "_node.md".

3. For each remaining file, exclude it if it lives
   inside a _-prefixed directory directly under
   "code-from-spec/":
     a. Remove the "code-from-spec/" prefix from the
        file path.
     b. Look for the first "/" in the remainder.
     c. If no "/" is found, the file is directly inside
        "code-from-spec/" — do not exclude it.
     d. If a "/" is found, extract the text before it
        as the first directory segment. If the first
        directory segment starts with "_", exclude this
        file. Otherwise, keep it.

4. For each file that was not excluded, call
   `LogicalNameFromPath` with the file's PathCfs to
   derive its logical name. If `LogicalNameFromPath`
   raises an error, propagate it. Build a SpecTreeNode
   record with: logical_name = the derived logical name,
   file_path = the file's PathCfs.

5. Sort all resulting SpecTreeNode records alphabetically
   by logical_name.

6. If the sorted list is empty, raise error
   "no nodes found".

7. Return the sorted list of SpecTreeNode records.

## Go-specific guidance

- Use the `listfiles` package for `ListFiles`.
- Use the `logicalnames` package for `LogicalNameFromPath`.
- Use the `pathutils` package for `PathCfs`.
- Extract the file name by finding the last `/` in the
  `PathCfs.Value` string.
- The package name should be `spectree`.
