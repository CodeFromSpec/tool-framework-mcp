---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/write_file(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_writer(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/logic/mcp_tools/write_file(write_file)
outputs:
  - id: mcpwritefile
    path: internal/mcpwritefile/mcpwritefile.go
---

# ROOT/golang/implementation/mcp_tools/write_file

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `logicalnames` package for `LogicalNameToPath`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `filewriter` package for `FileWrite`.
- The package name should be `mcpwritefile`.
- The function receives plain strings from the MCP
  transport layer. Construct `PathCfs` internally.
