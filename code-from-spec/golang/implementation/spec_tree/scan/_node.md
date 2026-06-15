---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/scan
  - ARTIFACT/golang/interfaces/os/list_files
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/logic/spec_tree/scan
output: internal/spectree/spectree.go
---

# SPEC/golang/implementation/spec_tree/scan

# Agent

Implement the pseudocode from the input as a Go package.
The output file is the sole .go file in the package —
declare all types, error sentinels, and function
signatures from the interface artifact in this file.

## Go-specific guidance

- Use the `listfiles` package for `ListFiles`.
- Use the `logicalnames` package for `LogicalNameFromPath`.
- Use the `pathutils` package for `PathCfs`.
- Extract the file name by finding the last `/` in the
  `PathCfs.Value` string.
- The package name should be `spectree`.
