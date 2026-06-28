---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/validate_specs
  - ARTIFACT/golang/interfaces/spec_tree/scan
  - ARTIFACT/golang/interfaces/spec_tree/validate
  - ARTIFACT/golang/interfaces/utils/node_ranking
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/parsing/artifact_tag
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/os/path_utils
output: internal/mcpvalidatespecs/mcpvalidatespecs_test.go
---

# SPEC/golang/tests/mcp_tools/validate_specs

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

## Test cases

### Happy path

#### Clean tree — no errors

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Compute the current chain hash for SPEC/a using
  `ChainHashCompute`. Create `out/a.go` with a valid
  artifact tag with that hash.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` is empty.
- `cycles` is empty.
- `staleness` is empty.

#### Stale artifact detected

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Create `out/a.go` with an artifact tag containing a
  27-character base64url string that differs from the
  current chain hash.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `staleness` contains one StalenessEntry for
  `"SPEC/a"` with `Status` = `"stale"` and `Rank`
  present.

#### Missing artifact detected

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Do not create `out/a.go`.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `staleness` contains one StalenessEntry for
  `"SPEC/a"` with `Status` = `"missing"`.

#### Malformed tag detected

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Create `out/a.go` with content that has no artifact
  tag.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `staleness` contains one StalenessEntry for
  `"SPEC/a"` with `Status` = `"malformed tag"`.

#### Staleness entries include rank

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  frontmatter `output: out/b.go`,
  `depends_on: ["SPEC/a"]`.
- Create `out/a.go` and `out/b.go` with outdated
  artifact tag hashes.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- Both StalenessEntries have `Rank` values.
- SPEC/a's rank is strictly less than SPEC/b's rank.

#### Staleness ordered by rank then name

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/z/_node.md` with `# SPEC/z`,
  frontmatter `output: out/z.go`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Both output files with outdated hashes.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- Staleness entries ordered: SPEC/a before SPEC/z
  (same rank, alphabetical).

### Format errors

#### Format error from invalid depends_on

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `depends_on: ["SPEC/missing"]`.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` contains a FormatError for
  `"SPEC/a"` with `Rule` = `"dependency_targets"`.

#### Format error from parse failure

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with invalid
  content (plain text before any heading).

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` contains a FormatError for
  `"SPEC/a"` with `Rule` = `"parse"`.

#### Continues after parse failure

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with invalid
  content.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  frontmatter `output: out/b.go`.
- Create `out/b.go` with outdated artifact tag hash.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` contains a FormatError for
  `"SPEC/a"`.
- `staleness` contains a StalenessEntry for
  `"SPEC/b"` with `Status` = `"stale"`.
- Both reported in the same report.

#### Subdirectory without _node.md detected

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`.
- Create empty directory `code-from-spec/b/` with no
  `_node.md`.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` contains a FormatError with
  `Rule` = `"missing_node_md"`.

#### _-prefixed dir under code-from-spec not flagged

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create directory `code-from-spec/_tools/` with no
  `_node.md`.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- No FormatError for `_tools/`.

### Cycle detection

#### Simple cycle detected

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `depends_on: ["SPEC/b"]`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  frontmatter `depends_on: ["SPEC/a"]`.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `cycles` is not empty, contains at least one of
  `"SPEC/a"` or `"SPEC/b"`.

#### Ranking skipped when format errors exist

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `depends_on: ["SPEC/missing"]`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  frontmatter `output: out/b.go`.
- Create `out/b.go` with outdated artifact tag hash.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` is not empty.
- Any StalenessEntry for `"SPEC/b"` has `Rank` = 0.

### Edge cases

#### Empty spec tree — scan fails

Setup:
- Do not create a `code-from-spec/` directory.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- `format_errors` contains a FormatError with
  `Rule` = `"scan"`.
- `cycles` is empty.
- `staleness` is empty.

#### Node with no output — not in staleness

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Context` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`.
  No output in frontmatter.

Actions:
1. Call `MCPValidateSpecs()`.

Expected:
- No StalenessEntry for `"SPEC/a"`.

## Go-specific guidance

- The package name is `mcpvalidatespecs_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
- Logical names map to filesystem paths:
  `SPEC` → `code-from-spec/_node.md`,
  `SPEC/x` → `code-from-spec/x/_node.md`.
- To compute a valid chain hash for clean-tree tests,
  use `ChainHashCompute` from the `chainhash` package.
- For stale artifacts, use any 27-character base64url
  string that differs from the current chain hash.
