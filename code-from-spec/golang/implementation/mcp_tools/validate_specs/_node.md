---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/validate_specs(interface)
  - ARTIFACT/golang/interfaces/spec_tree/scan(interface)
  - ARTIFACT/golang/interfaces/spec_tree/validate(interface)
  - ARTIFACT/golang/interfaces/utils/node_ranking(interface)
  - ARTIFACT/golang/interfaces/chain/resolver(interface)
  - ARTIFACT/golang/interfaces/chain/hash(interface)
  - ARTIFACT/golang/interfaces/parsing/artifact_tag(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/logic/mcp_tools/validate_specs(validate_specs)
outputs:
  - id: mcpvalidatespecs
    path: internal/mcpvalidatespecs/mcpvalidatespecs.go
---

# ROOT/golang/implementation/mcp_tools/validate_specs

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `spectree` package for `SpecTreeScan`.
- Use the `spectreevalidate` package for
  `SpecTreeValidate` and `SpecTreeValidateInput`,
  `FormatError`.
- Use the `noderanking` package for `NodeRankCompute`,
  `NodeRankInput`, `NodeRankEntry`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `artifacttag` package for `ArtifactTagExtract`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `parsenode` package for `NodeParse`.
- Use the `pathutils` package for `PathValidateCfs`,
  `PathCfs`.
- The package name should be `mcpvalidatespecs`.
- `StalenessEntry`, `ValidationReport` are exported
  structs.
- The function never returns an error — all problems
  are collected in the report.
