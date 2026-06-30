---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - ARTIFACT/domain/code-from-spec/manifest-format
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcploadchain/mcploadchain_test.go
---

# SPEC/golang/test/cases/mcp_tools/load_chain

# Agent

## Test setup guidance

`MCPLoadChain` calls `ChainResolve`, `ChainHashCompute`,
`parsing.ParseNode`, `manifest.OpenManifest`, and
`oslayer.OpenFile`
internally. Tests must create a complete spec tree on
disk with valid `_node.md` files. Use `testutils.Chdir` and
create `code-from-spec/.../_node.md` files with
frontmatter and body content matching the test setup.

Node files must have valid structure for
`parsing.ParseNode`:
at minimum a `# <logical_name>` heading as the first
heading. Leaf nodes need frontmatter with `output`.

For ARTIFACT and external file tests, create the
referenced files on disk at the declared paths.

The output is an XML document with `<chain>` as root
element containing `<existing_artifact>`,
`<constraints>`, `<instructions>`, and `<input>`
sections. No `chain_hash:` prefix line.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- Output starts with `<chain>`.
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
  `# SPEC/root/a/b`, frontmatter `output: out/b.txt`,
  `# Public` → `## Contract` with content.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a/b")`.

Expected:
- `<constraints>` contains three entries:
  `<entry name="SPEC/root">` with `## Overview`,
  `<entry name="SPEC/root/a">` with `## Details`,
  `<entry name="SPEC/root/a/b">` with `## Contract`.

#### Ancestor without public section skipped

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root` heading only (no public section).
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, `# Public` → `## Interface` with
  content, frontmatter `output: out/a.txt`.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains
  `<entry name="SPEC/root/b(interface)">` with
  `## Interface` content only. Does not contain
  `## Constraints`.

#### ARTIFACT dependency — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/b.go`.
- Create `out/b.go` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`,
  `depends_on: ["ARTIFACT/root/b"]`.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- `<constraints>` contains
  `<entry name="ARTIFACT/root/b">` with the full
  content of `out/b.go`.

#### EXTERNAL dependency — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `data/config.yaml` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `depends_on: ["EXTERNAL/data/config.yaml"]`.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- No `<instructions>` element in the output.

#### Input present — ARTIFACT

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `output: out/data.json`.
- Create `out/data.json` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.txt`,
  `input: ARTIFACT/root/b`.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- `<input>` contains the full content of
  `out/data.json`.
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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")` twice.

Expected:
- Both calls return identical output strings.

### Error cases

#### Invalid logical name — not SPEC/

Actions:
1. Call `mcploadchain.MCPLoadChain("INVALID/something")`.

Expected:
- Returns error `parsing.ErrUnrecognizedPrefix`.

#### Nonexistent node file

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/nonexistent")` with no
   `_node.md` on disk.

Expected:
- Returns error propagated from `parsing.ParseNode`
  (`oslayer.ErrFileUnreadable`).

#### No output declared

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No output in frontmatter.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns error `mcploadchain.ErrNoOutput`.

#### Invalid output path — traversal

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter
  `output: ../../etc/passwd`.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns error `mcploadchain.ErrInvalidOutputPath`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns error `mcploadchain.ErrArtifactModified`.

#### No manifest — modified check skipped

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with known content.
- No `.manifest` file.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

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
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")`.

Expected:
- Returns an error — the missing node is detected
  during chain processing.

## Go-specific guidance

- The package name is `mcploadchain_test` (external
  test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
- To verify XML output, use `strings.Contains` to
  check for expected elements and content. Do not
  parse with `encoding/xml` — simple string checks
  are sufficient.
- The manifest file path is `code-from-spec/.manifest`
  — write manifest fixtures there, not at `.manifest`
  in the working directory root.
