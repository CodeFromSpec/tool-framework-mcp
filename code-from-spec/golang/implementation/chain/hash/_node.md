---
depends_on:
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/utils/text_normalization
input: ARTIFACT/functional/logic/chain/hash
output: internal/chainhash/chainhash.go
---

# ROOT/golang/implementation/chain/hash

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `chainresolver` package for the `Chain` and
  `ChainItem` records.
- Use the `parsenode` package for `NodeParse` and the
  `Node`, `NodeSection`, `NodeSubsection` records.
- Use the `filereader` package for `FileOpen`,
  `FileReadLine`, `FileSkipLines`, `FileClose`.
- Use the `pathutils` package for `PathCfs`.
- Use the `logicalnames` package for
  `LogicalNameIsArtifact`.
- Use the `textnormalization` package for `NormalizeText`.
- Use the `frontmatter` package for `FrontmatterExternal`.
- For SHA-1 and base64url, use `crypto/sha1` and
  `encoding/base64` (base64.RawURLEncoding).
- The package name should be `chainhash`.
