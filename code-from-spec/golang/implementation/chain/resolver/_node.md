---
depends_on:
  - ARTIFACT/golang/interfaces/chain/resolver(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/logic/chain/resolver(chain_resolver)
outputs:
  - id: chainresolver
    path: internal/chainresolver/chainresolver.go
---

# ROOT/golang/implementation/chain/resolver

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `logicalnames` package for `LogicalNameGetParent`,
  `LogicalNameToPath`, `LogicalNameGetQualifier`,
  `LogicalNameStripQualifier`, `LogicalNameGetArtifactGenerator`,
  `LogicalNameIsArtifact`.
- Use the `frontmatter` package for `FrontmatterParse` and
  the `Frontmatter`, `FrontmatterExternal`,
  `FrontmatterOutput` records.
- Use the `pathutils` package for `PathCfs`.
- The package name should be `chainresolver`.
- `ChainItem` and `Chain` are exported structs in this package.
- `FrontmatterExternal` in the `Chain.External` field uses
  the type from the `frontmatter` package directly.
