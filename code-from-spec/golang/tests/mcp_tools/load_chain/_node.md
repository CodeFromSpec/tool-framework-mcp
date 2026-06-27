---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/tests/mcp_tools/load_chain
output: internal/mcploadchain/mcploadchain_test.go
---

# SPEC/golang/tests/mcp_tools/load_chain

# Agent

## Test setup guidance

`MCPLoadChain` calls `ChainResolve`, `ChainHashCompute`,
`NodeParse`, `FrontmatterParse`, and `FileOpen`
internally. Tests must create a complete spec tree on
disk with valid `_node.md` files. Use `testChdir` and
create `code-from-spec/.../_node.md` files with
frontmatter and body content matching the test setup.

Node files must have valid structure for `NodeParse`:
at minimum a `# <logical_name>` heading as the first
heading. Leaf nodes need frontmatter with `output`.

For ARTIFACT and external file tests, create the
referenced files on disk at the declared paths.
