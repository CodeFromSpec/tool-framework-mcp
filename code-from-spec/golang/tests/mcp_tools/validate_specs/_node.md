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
input: ARTIFACT/functional/tests/mcp_tools/validate_specs(validate_specs_tests)
outputs:
  - id: mcpvalidatespecs_test
    path: internal/mcpvalidatespecs/mcpvalidatespecs_test.go
---

# ROOT/golang/tests/mcp_tools/validate_specs

# Agent

## Test setup guidance

`MCPValidateSpecs` calls `SpecTreeScan`, `NodeParse`,
`FrontmatterParse`, `SpecTreeValidate`,
`NodeRankCompute`, `ChainResolve`, `ChainHashCompute`,
and `ArtifactTagExtract` internally. Tests must create
a complete spec tree on disk.

Use `testChdir` and create `code-from-spec/.../_node.md`
files with valid structure (frontmatter + body with
`# <logical_name>` heading).

For staleness tests, create output files with artifact
tags. To produce a matching hash, call `MCPValidateSpecs`
once to discover the current chain hash, then write an
artifact tag with that hash.

The function never returns an error — always check the
fields of the returned `ValidationReport`.
