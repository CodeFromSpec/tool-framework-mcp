---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/chain_hash(interface)
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain(interface)
  - ARTIFACT/golang/interfaces/chain/resolver(interface)
  - ARTIFACT/golang/interfaces/chain/hash(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/mcp_tools/chain_hash(chain_hash_tests)
outputs:
  - id: mcpchainhash_test
    path: internal/mcpchainhash/mcpchainhash_test.go
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
