---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/chain_hash
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/logic/mcp_tools/chain_hash
output: internal/mcpchainhash/mcpchainhash.go
---

# SPEC/golang/implementation/mcp_tools/chain_hash

# Agent

## Go-specific guidance

- The package name is `mcpchainhash`.
- Import `chainresolver`, `chainhash`, `frontmatter`,
  `logicalnames`, and `pathutils` packages.
- The function is simple: resolve, parse frontmatter,
  check output exists, resolve chain, compute hash,
  return.
