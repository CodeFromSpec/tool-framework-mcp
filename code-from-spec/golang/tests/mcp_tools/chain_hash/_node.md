---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/chain_hash
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/tests/mcp_tools/chain_hash
output: internal/mcpchainhash/mcpchainhash_test.go
---

# ROOT/golang/tests/mcp_tools/chain_hash

# Agent

## Go-specific guidance

- The package name is `mcpchainhash_test` (external test
  package).
- The "hash matches load_chain hash" test imports
  `mcploadchain` to call `MCPLoadChain` and compare
  the chain_hash field.
- Use `testChdir` and `testWriteFile` helpers for
  creating spec trees on disk.
