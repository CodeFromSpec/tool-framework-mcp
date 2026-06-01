---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain(interface)
  - ARTIFACT/golang/interfaces/chain/resolver(interface)
  - ARTIFACT/golang/interfaces/chain/hash(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/mcp_tools/load_chain(load_chain_tests)
outputs:
  - id: mcploadchain_test
    path: internal/mcploadchain/mcploadchain_test.go
---

# ROOT/golang/tests/mcp_tools/load_chain

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
heading. Leaf nodes need frontmatter with `outputs`.

For ARTIFACT and external file tests, create the
referenced files on disk at the declared paths.
