---
depends_on:
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/logic/chain/resolver
output: internal/chainresolver/chainresolver.go
---

# SPEC/golang/implementation/chain/resolver

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
