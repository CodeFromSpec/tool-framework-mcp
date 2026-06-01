---
depends_on:
  - ARTIFACT/golang/interfaces/chain/hash(interface)
  - ARTIFACT/golang/interfaces/chain/resolver(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
  - ARTIFACT/golang/interfaces/utils/text_normalization(interface)
input: ARTIFACT/functional/logic/chain/hash(chain_hash)
outputs:
  - id: chainhash
    path: internal/chainhash/chainhash.go
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
- Use the `frontmatter` package for `FrontmatterExternal`
  and `FrontmatterExternalFragment`.
- For SHA-1 and base64url, use `crypto/sha1` and
  `encoding/base64` (base64.RawURLEncoding).
- The package name should be `chainhash`.
