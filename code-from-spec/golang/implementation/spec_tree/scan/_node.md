---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/scan(interface)
  - ARTIFACT/golang/interfaces/os/list_files(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/logic/spec_tree/scan(spec_tree)
outputs:
  - id: spectree
    path: internal/spectree/spectree.go
---

# ROOT/golang/implementation/spec_tree/scan

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `listfiles` package for `ListFiles`.
- Use the `logicalnames` package for `LogicalNameFromPath`.
- Use the `pathutils` package for `PathCfs`.
- Extract the file name by finding the last `/` in the
  `PathCfs.Value` string.
- The package name should be `spectree`.
