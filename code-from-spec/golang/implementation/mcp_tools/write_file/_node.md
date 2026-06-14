---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/write_file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/file_writer
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/logic/mcp_tools/write_file
output: internal/mcpwritefile/mcpwritefile.go
---

# SPEC/golang/implementation/mcp_tools/write_file

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
