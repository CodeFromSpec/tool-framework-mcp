---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/chain_hash(interface)
  - ARTIFACT/golang/interfaces/chain/resolver(interface)
  - ARTIFACT/golang/interfaces/chain/hash(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/logic/mcp_tools/chain_hash(chain_hash)
outputs:
  - id: mcpchainhash
    path: internal/mcpchainhash/mcpchainhash.go
---

# ROOT/golang/implementation/mcp_tools/chain_hash

# Agent

## Go-specific guidance

- The package name is `mcpchainhash`.
- Import `chainresolver`, `chainhash`, `frontmatter`,
  `logicalnames`, and `pathutils` packages.
- The function is simple: resolve, parse frontmatter,
  check output exists, resolve chain, compute hash,
  return.
