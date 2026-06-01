---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/write_file(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_writer(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/mcp_tools/write_file(write_file_tests)
outputs:
  - id: mcpwritefile_test
    path: internal/mcpwritefile/mcpwritefile_test.go
---

# ROOT/golang/tests/mcp_tools/write_file

# Agent

## Test setup guidance

`MCPWriteFile` reads the node's frontmatter from disk
to validate the path against declared outputs. Tests
must create `_node.md` files with frontmatter containing
outputs declarations. Use `testChdir` and create the
spec tree structure (`code-from-spec/.../_node.md`).
