---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/write_file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/file_writer
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/tests/mcp_tools/write_file
output: internal/mcpwritefile/mcpwritefile_test.go
---

# ROOT/golang/tests/mcp_tools/write_file

# Agent

## Test setup guidance

`MCPWriteFile` reads the node's frontmatter from disk
to validate the path against declared outputs. Tests
must create `_node.md` files with frontmatter containing
outputs declarations. Use `testChdir` and create the
spec tree structure (`code-from-spec/.../_node.md`).
