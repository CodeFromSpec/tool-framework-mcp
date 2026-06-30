---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/mcp_tools/validate_specs
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/spec_tree/scan
  - SPEC/golang/implementation/spec_tree/validate
  - SPEC/golang/implementation/spec_tree/ranking
output: internal/mcpvalidatespecs/mcpvalidatespecs_test.go
---

# SPEC/golang/test/cases/mcp_tools/validate_specs

# Agent

## Test setup guidance

`MCPValidateSpecs` calls `SpecTreeScan`,
`parsing.ParseNode`, `SpecTreeValidate`,
`NodeRankCompute`, `ChainResolve`, `ChainHashCompute`,
and `manifest.OpenManifest` internally. Tests must create a
complete spec tree on disk.

Use `testutils.Chdir` and create `code-from-spec/.../_node.md`
files with valid structure (frontmatter + body with
`# <logical_name>` heading).

For staleness tests, create a `.manifest` file with
the appropriate entries. To produce a matching chain
hash for clean-tree tests, call `ChainHashCompute`.
For the file checksum, compute SHA-1 of the file
content (base64url, 27 chars).

The function never returns an error — always check the
fields of the returned `mcpvalidatespecs.ValidationReport`.

## Test cases

### Happy path

#### Clean tree — no errors

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with known content.
- Compute the current chain hash for SPEC/root/a.
  Compute the checksum of `out/a.go`.
- Create `code-from-spec/.manifest` with header
  `code-from-spec: v5` and entry:
  `ARTIFACT/root/a;path:out/a.go;checksum:<checksum>;chain:<chain_hash>`

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` is empty.
- `cycles` is empty.
- `staleness` is empty.

#### Stale artifact detected

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with known content.
- Create `.manifest` with a chain hash that differs
  from the current chain hash (but checksum matches
  the file).

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `staleness` contains one mcpvalidatespecs.StalenessEntry for
  `"SPEC/root/a"` with `Status` = `"stale"` and
  `Rank` present.

#### Missing artifact — no manifest entry

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- No manifest entry for ARTIFACT/root/a. No file on
  disk.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `staleness` contains one mcpvalidatespecs.StalenessEntry for
  `"SPEC/root/a"` with `Status` = `"missing"`.

#### Missing artifact — file does not exist

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `.manifest` with a valid entry for
  ARTIFACT/root/a (matching chain hash), but do not
  create `out/a.go` on disk.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `staleness` contains one mcpvalidatespecs.StalenessEntry for
  `"SPEC/root/a"` with `Status` = `"missing"`.

#### Modified artifact detected

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "original".
- Create `.manifest` with chain hash matching current,
  but checksum matching the hash of "original".
- Overwrite `out/a.go` with content "modified" (so
  file hash no longer matches manifest checksum).

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `staleness` contains one mcpvalidatespecs.StalenessEntry for
  `"SPEC/root/a"` with `Status` = `"modified"`.

#### Orphan manifest entry detected

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `.manifest` with entry for
  `ARTIFACT/root/deleted` with path `out/deleted.go`.
- No `code-from-spec/root/deleted/_node.md` on disk.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `staleness` contains one mcpvalidatespecs.StalenessEntry with
  `Status` = `"orphan"`.

#### Staleness entries include rank

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/b.go`,
  `depends_on: ["SPEC/root/a"]`.
- No manifest entries (both are missing).

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- Both StalenessEntries have `Rank` values.
- SPEC/root/a's rank is strictly less than
  SPEC/root/b's rank.

#### Staleness ordered by rank then name

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/z/_node.md` with
  `# SPEC/root/z`, frontmatter `output: out/z.go`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- No manifest entries (both are missing).

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- Staleness entries ordered: SPEC/root/a before
  SPEC/root/z (same rank, alphabetical).

### Format errors

#### Format error from invalid depends_on

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`,
  frontmatter `depends_on: ["SPEC/root/missing"]`.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` contains a spectreevalidate.FormatError for
  `"SPEC/root/a"` with `Rule` = `"dependency_targets"`.

#### Format error from parse failure

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with invalid
  content (plain text before any heading).

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` contains a spectreevalidate.FormatError for
  `"SPEC/root/a"` with `Rule` = `"parse"`.

#### Continues after parse failure

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with invalid
  content.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/b.go`.
- No manifest entry for ARTIFACT/root/b.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` contains a spectreevalidate.FormatError for
  `"SPEC/root/a"`.
- `staleness` contains a mcpvalidatespecs.StalenessEntry for
  `"SPEC/root/b"` with `Status` = `"missing"`.
- Both reported in the same report.

#### Subdirectory without _node.md detected

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`.
- Create empty directory `code-from-spec/root/b/` with
  no `_node.md`.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` contains a spectreevalidate.FormatError with
  `Rule` = `"missing_node_md"`.

#### .-prefixed dir under code-from-spec not flagged

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create directory `code-from-spec/.cache/` with no
  `_node.md`.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- No spectreevalidate.FormatError for `.cache/`.

### Cycle detection

#### Simple cycle detected

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`,
  frontmatter `depends_on: ["SPEC/root/b"]`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`,
  frontmatter `depends_on: ["SPEC/root/a"]`.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `cycles` is not empty, contains at least one of
  `"SPEC/root/a"` or `"SPEC/root/b"`.

#### Ranking skipped when format errors exist

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`,
  frontmatter `depends_on: ["SPEC/root/missing"]`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/b.go`.
- No manifest entry for ARTIFACT/root/b.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` is not empty.
- Any mcpvalidatespecs.StalenessEntry for `"SPEC/root/b"` has
  `Rank` = 0.

### Edge cases

#### Empty spec tree — scan fails

Setup:
- Do not create a `code-from-spec/` directory.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `format_errors` contains a spectreevalidate.FormatError with
  `Rule` = `"scan"`.
- `cycles` is empty.
- `staleness` is empty.

#### Node with no output — not in staleness

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No output in frontmatter.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- No mcpvalidatespecs.StalenessEntry for `"SPEC/root/a"`.

#### No manifest file — all artifacts with output are missing

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Do not create `.manifest`.

Actions:
1. Call `mcpvalidatespecs.MCPValidateSpecs()`.

Expected:
- `staleness` contains one mcpvalidatespecs.StalenessEntry for
  `"SPEC/root/a"` with `Status` = `"missing"`.

## Go-specific guidance

- The package name is `mcpvalidatespecs_test` (external
  test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
- Logical names map to filesystem paths:
  `SPEC/x` → `code-from-spec/x/_node.md`.
- To compute a valid chain hash for clean-tree tests,
  use `ChainHashCompute` from the `chainhash` package.
- To compute a file checksum, use SHA-1 of the file
  content (after CRLF→LF normalization, with trailing
  LF), encoded as base64url (27 chars).
- Create `.manifest` files using
  `manifest.OpenManifest(false)` + `m.Save()`, or by
  writing the file directly
  with the correct format:
  ```
  code-from-spec: v5
  ARTIFACT/root/a;path:out/a.go;checksum:<hash>;chain:<hash>
  ```
