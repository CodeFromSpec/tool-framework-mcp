---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/utils/text_normalization
input: ARTIFACT/functional/logic/mcp_tools/load_chain
output: internal/mcploadchain/mcploadchain.go
---

# ROOT/golang/implementation/mcp_tools/load_chain

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `chainresolver` package for `ChainResolve` and
  the `Chain`, `ChainItem` records.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `parsenode` package for `NodeParse` and the
  `Node`, `NodeSection`, `NodeSubsection` records.
- Use the `filereader` package for `FileOpen`,
  `FileReadLine`, `FileSkipLines`, `FileClose`.
- Use the `frontmatter` package for `FrontmatterParse`
  and the `Frontmatter`, `FrontmatterExternal` records.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `logicalnames` package for `LogicalNameToPath`
  and `LogicalNameIsArtifact`.
- Use the `textnormalization` package for `NormalizeText`.
- The package name should be `mcploadchain`.
- `MCPLoadChainResult` is an exported struct.
- When reconstructing content from lines, append `\n`
  after each line including the last.
