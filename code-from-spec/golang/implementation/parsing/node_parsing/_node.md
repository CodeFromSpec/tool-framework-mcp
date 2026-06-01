---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
  - ARTIFACT/golang/interfaces/utils/text_normalization(interface)
input: ARTIFACT/functional/logic/parsing/node_parsing(node_parsing)
outputs:
  - id: parsenode
    path: internal/parsenode/parsenode.go
---

# ROOT/golang/implementation/parsing/node_parsing

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `textnormalization.NormalizeText` for all heading
  comparisons.
- Use `logicalnames.LogicalNameToPath` to resolve logical
  names to file paths.
- Use the `filereader` package for file I/O: `FileOpen`,
  `FileReadLine`, `FileClose`.
- The package name should be `parsenode`.
