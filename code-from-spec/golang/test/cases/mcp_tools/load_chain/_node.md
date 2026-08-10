---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - ARTIFACT/external/code-from-spec/manifest-format
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcploadchain/mcploadchain_test.go
---

# SPEC/golang/test/cases/mcp_tools/load_chain

# Agent

## Test setup guidance

`MCPLoadChain` takes an opaque token, not a raw logical
name. Tests must first call
`subagenttoken.SubagentTokenGenerate(logicalName)` to
obtain a token, then pass that token to `MCPLoadChain`.

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
`<constraints>`, `<references>`, `<instructions>`,
and `<input>` sections. No `chain_hash:` prefix line.

## Test cases

### Happy path

#### Simple leaf node — constraints and hash

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root` heading, `# Public` with `## Context`
  subsection containing one line of content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a` heading, frontmatter
  `type: artifact`, `output: out/a.txt`, `# Public` with `## Interface`
  subsection, `# Agent` section with content.
- Do not create `out/a.txt`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

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
  `# SPEC/root/a/b`, frontmatter `type: artifact`, `output: out/b.txt`,
  `# Public` → `## Contract` with content.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a/b")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

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
  content, frontmatter `type: artifact`, `output: out/a.txt`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

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
  content, frontmatter `type: artifact`, `output: out/a.txt`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

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
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `imports: ["SPEC/root/b"]`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<references>` contains
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
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `imports: ["SPEC/root/b(interface)"]`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<references>` contains
  `<entry name="SPEC/root/b(interface)">` with
  `## Interface` content only. Does not contain
  `## Constraints`.

#### ARTIFACT dependency — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `type: artifact`, `output: out/b.go`.
- Create `out/b.go` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`,
  `imports: ["ARTIFACT/root/b"]`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<references>` contains
  `<entry name="ARTIFACT/root/b">` with the full
  content of `out/b.go`.

#### EXTERNAL dependency — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `data/config.yaml` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `imports: ["EXTERNAL/data/config.yaml"]`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<references>` contains
  `<entry name="EXTERNAL/data/config.yaml">` with
  the full content of `data/config.yaml`.

#### Target agent section in instructions

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `# Public` → `## Interface` with content,
  `# Agent` with content.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<constraints>` contains target's `## Interface`.
- `<instructions>` contains agent content without
  `# Agent` heading.

#### Target without agent section — no instructions

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `# Public` → `## Interface` with content. No
  `# Agent` section.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- No `<instructions>` element in the output.

#### Input present — ARTIFACT

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `type: artifact`, `output: out/data.json`.
- Create `out/data.json` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `input: ARTIFACT/root/b` (scalar form).

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<input>` contains `<entry name="ARTIFACT/root/b">`
  with the full content of `out/data.json`.
- Input content does not appear in `<constraints>`.

#### EXTERNAL input — full content

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `docs/vendor/spec.yaml` with known content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `input: EXTERNAL/docs/vendor/spec.yaml` (scalar form).

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<input>` contains
  `<entry name="EXTERNAL/docs/vendor/spec.yaml">` with
  the full content of `docs/vendor/spec.yaml`.

#### SPEC input — public content extracted

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, `# Public` → `## Acceptance tests`
  with content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `input: SPEC/root/b` (scalar form).

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<input>` contains `<entry name="SPEC/root/b">` with
  `## Acceptance tests` content from SPEC/root/b.

#### Multiple inputs — each own entry

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/b/_node.md` with
  `# SPEC/root/b`, frontmatter `type: artifact`, `output: out/b.json`.
- Create `out/b.json` with known content.
- Create `code-from-spec/root/c/_node.md` with
  `# SPEC/root/c`, `# Public` → `## Acceptance tests`
  with content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `input:\n    - ARTIFACT/root/b\n    - SPEC/root/c`
  (list form).

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- A single `<input>` block contains two entries:
  `<entry name="ARTIFACT/root/b">` with the content of
  `out/b.json`, and `<entry name="SPEC/root/c">` with
  the `## Acceptance tests` content.

#### No input — section absent

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`.
  No input field.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- No `<input>` element in output.

#### Existing artifact present

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with known content.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- `<existing_artifact>` contains the full content
  of `out/a.go`.

#### Existing artifact absent — section omitted

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Do not create `out/a.go`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- No `<existing_artifact>` element in output.

#### Hash is deterministic

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Overview` with
  stable content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)` twice, reusing the
   same `token`.

Expected:
- Both calls return identical output strings.

### Error cases

#### Invalid token — malformed

Actions:
1. Call `mcploadchain.MCPLoadChain("not-a-valid-token")`.

Expected:
- Returns error `subagenttoken.ErrInvalidToken`.

#### Invalid logical name — not SPEC/

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("INVALID/something")`
   → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Returns error `parsing.ErrUnrecognizedPrefix`.

#### Nonexistent node file

Setup:
- Do not create `code-from-spec/root/nonexistent/_node.md`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/nonexistent")`
   → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Returns error propagated from `parsing.ParseNode`
  (`oslayer.ErrFileUnreadable`).

#### No type declared

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No type in frontmatter.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Returns error `mcploadchain.ErrNoOutput`.

#### Invalid output path — traversal

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter
  `type: artifact`, `output: ../../etc/passwd`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Returns error `mcploadchain.ErrInvalidOutputPath`.

#### Modified artifact blocked

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with content "original".
- Create `.manifest` with entry for ARTIFACT/root/a
  with checksum matching "original" and a valid chain
  hash.
- Overwrite `out/a.go` with content "modified" (file
  hash no longer matches manifest checksum).

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Returns error `mcploadchain.ErrModified`.

#### No manifest — modified check skipped

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with known content.
- No `.manifest` file.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- No error. Chain is loaded normally. The modified
  check is skipped when no manifest exists.

#### Unresolvable dependency

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.txt`,
  `imports: ["SPEC/root/missing"]`.
- Do not create `code-from-spec/root/missing/_node.md`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Returns an error — the missing node is detected
  during chain processing.

### Verdict chains

#### Verdict chain — no existing artifact section

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with content.
- Create `code-from-spec/root/v/_node.md` with
  `# SPEC/root/v`, frontmatter `type: verdict`,
  `output: code-from-spec/root/v/verdict.md`.
  `# Agent` section with instructions.
- Create the output file
  `code-from-spec/root/v/verdict.md` on disk with
  previous content.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/v")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- No error.
- Output does NOT contain `<existing_artifact>`.
- Output does NOT contain `<previous_constraints>`.
- Output does NOT contain `<previous_references>`.
- Output does NOT contain `<previous_instructions>`.
- Output does NOT contain `<previous_input>`.
- Output does NOT contain `disposition=`.
- Output DOES contain `<constraints>` with the ancestor
  and target public content.
- Output DOES contain `<instructions>` with the agent
  content.

#### Verdict chain — no dispositions even with manifest

Setup:
- Same as above, plus create `.manifest` with a
  `VERDICT/root/v` entry whose checksum matches the
  SHA-1 hash of the verdict file content on disk
  (computed with CRLF→LF normalization and trailing
  LF, encoded as base64url 27 chars). Use any value
  for chain hash. Set `Result` = `"pass"`.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/v")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- No error.
- Output does NOT contain `disposition=`.
- No `<previous_*>` or `<existing_artifact>` sections.

#### Modified verdict — blocked

Setup:
- Create spec tree and verdict file.
- Create `.manifest` with `VERDICT/root/v` entry where
  checksum does NOT match the file on disk.

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/v")` → `token`.
2. Call `mcploadchain.MCPLoadChain(token)`.

Expected:
- Error `mcploadchain.ErrModified`.

## Go-specific guidance

- The package name is `mcploadchain_test` (external
  test package).
- Import the `subagenttoken` package to mint tokens for
  `MCPLoadChain` calls.
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
