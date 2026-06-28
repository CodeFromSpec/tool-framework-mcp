---
depends_on:
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/parsing/node_parsing
  - SPEC/golang/implementation/utils/logical_names
output: internal/mcploadchain/mcploadchain_test.go
---

# SPEC/golang/tests/mcp_tools/load_chain

# Agent

## Test setup guidance

`MCPLoadChain` calls `ChainResolve`, `ChainHashCompute`,
`NodeParse`, `FrontmatterParse`, `ManifestOpen`, and
`FileOpen` internally. Tests must create a complete spec
tree on disk with valid `_node.md` files. Use `testChdir`
and create `code-from-spec/.../_node.md` files with
frontmatter and body content matching the test setup.

Node files must have valid structure for `NodeParse`:
at minimum a `# <logical_name>` heading as the first
heading. Leaf nodes need frontmatter with `output`.

For ARTIFACT and external file tests, create the
referenced files on disk at the declared paths.

The output format is: first line `chain_hash: <hash>`,
followed by an XML document with `<chain>` as root
element containing `<existing_artifact>`,
`<constraints>`, `<instructions>`, and `<input>`
sections.

## Test cases

### Happy path

#### Simple leaf node — constraints and hash

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root` heading, `# Public` with `## Context`
  subsection containing one line of content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a` heading, frontmatter
  `output: out/a.txt`, `# Public` with `## Interface`
  subsection, `# Agent` section with content.
- Do not create `out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- First line matches `chain_hash: ` followed by
  exactly 27 non-whitespace characters.
- Contains `<chain>` root element.
- `<constraints>` contains `<entry name="SPEC/root">`
  with `## Context` content, and
  `<entry name="SPEC/root/a">` with `## Interface`
  content. No `# Public` headings appear.
- `<instructions>` contains the agent content
  (without `# Agent` heading).
- No `<existing_artifact>` section.
- No `<input>` section.

#### Ancestor public content included

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Overview` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, `# Public` → `## Details` with
  content.
- Create `code-from-spec/root/a/b/_node.md` with
  `# SPEC/root/a/b`, frontmatter `output: out/b.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a/b")`.

Expected:
- `<constraints>` contains
  `<entry name="SPEC/root">` with `## Overview`,
  `<entry name="SPEC/root/a">` with `## Details`,
  and `<entry name="SPEC/root/a/b">` (if it has
  public content).

#### Ancestor without public section skipped

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root` heading only (no public section).
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, `# Public` → `## Interface` with
  content, frontmatter `output: out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` does not contain an entry for
  SPEC/root. Contains `<entry name="SPEC/root/a">`
  with `## Interface` content.

#### Ancestor with empty public section skipped

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` present but empty (no
  subsections).
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, `# Public` → `## Interface` with
  content, frontmatter `output: out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` does not contain an entry for
  SPEC/root.

#### Dependency without qualifier — public included

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, `# Public` → `## Interface` +
  `## Constraints`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/root/b"]`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains
  `<entry name="SPEC/root/b">` with `## Interface`
  and `## Constraints` content.

#### Dependency with qualifier — subsection only

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, `# Public` → `## Interface` +
  `## Constraints`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/root/b(interface)"]`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains
  `<entry name="SPEC/root/b(interface)">` with
  `## Interface` content only. Does not contain
  `## Constraints`.

#### ARTIFACT dependency — artifact tag line removed

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/b.go`.
- Create `out/b.go` with artifact tag line and body
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`,
  `depends_on: ["ARTIFACT/root/b"]`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains
  `<entry name="ARTIFACT/root/b">` with body content
  of `out/b.go` but without the artifact tag line.

#### EXTERNAL dependency — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `data/config.yaml` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `depends_on: ["EXTERNAL/data/config.yaml"]`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains
  `<entry name="EXTERNAL/data/config.yaml">` with
  the full content of `data/config.yaml`.

#### Target agent section in instructions

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `# Public` → `## Interface` with content,
  `# Agent` with content.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains target's `## Interface`.
- `<instructions>` contains agent content without
  `# Agent` heading.

#### Target without agent section — no instructions

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `# Public` → `## Interface` with content. No
  `# Agent` section.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- No `<instructions>` element in the output.

#### Input present — ARTIFACT

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/data.json`.
- Create `out/data.json` with artifact tag line and
  body content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `input: ARTIFACT/root/b`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<input>` contains body of `out/data.json` without
  artifact tag line.
- Input content does not appear in `<constraints>`.

#### EXTERNAL input — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `docs/vendor/spec.yaml` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `input: EXTERNAL/docs/vendor/spec.yaml`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<input>` contains the full content of
  `docs/vendor/spec.yaml`.

#### SPEC input — public content extracted

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, `# Public` → `## Acceptance tests`
  with content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `input: SPEC/root/b`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<input>` contains `## Acceptance tests` content
  from SPEC/root/b.

#### No input — section absent

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`.
  No input field.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- No `<input>` element in output.

#### Existing artifact present

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with known content.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- `<existing_artifact>` contains the full content
  of `out/a.go`.

#### Existing artifact absent — section omitted

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Do not create `out/a.go`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- No `<existing_artifact>` element in output.

#### Hash is deterministic

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Overview` with
  stable content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")` twice.

Expected:
- Both calls return identical `chain_hash` values.

### Error cases

#### Invalid logical name — not SPEC/

Actions:
1. Call `MCPLoadChain("INVALID/something")`.

Expected:
- Returns error `logicalnames.ErrUnrecognizedPrefix`.

#### Nonexistent node file

Actions:
1. Call `MCPLoadChain("SPEC/root/nonexistent")` with no
   `_node.md` on disk.

Expected:
- Returns error propagated from `FrontmatterParse`
  (`file.ErrFileUnreadable`).

#### No output declared

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No output in frontmatter.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns error `ErrNoOutput`.

#### Invalid output path — traversal

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter
  `output: ../../etc/passwd`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns error `ErrInvalidOutputPath`.

#### Modified artifact blocked

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "original".
- Create `.manifest` with entry for ARTIFACT/root/a
  with checksum matching "original" and a valid chain
  hash.
- Overwrite `out/a.go` with content "modified" (file
  hash no longer matches manifest checksum).

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns error `ErrArtifactModified`.

#### No manifest — modified check skipped

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with known content.
- No `.manifest` file.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- No error. Chain is loaded normally. The modified
  check is skipped when no manifest exists.

#### Unresolvable dependency

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/root/missing"]`.
- Do not create `code-from-spec/root/missing/_node.md`.

Actions:
1. Call `MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns an error — the missing node is detected
  during chain processing.

## Go-specific guidance

- The package name is `mcploadchain_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
- To verify XML output, use `strings.Contains` to
  check for expected elements and content. Do not
  parse with `encoding/xml` — simple string checks
  are sufficient.
