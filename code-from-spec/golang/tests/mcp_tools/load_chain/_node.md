---
depends_on:
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
`NodeParse`, `FrontmatterParse`, and `FileOpen`
internally. Tests must create a complete spec tree on
disk with valid `_node.md` files. Use `testChdir` and
create `code-from-spec/.../_node.md` files with
frontmatter and body content matching the test setup.

Node files must have valid structure for `NodeParse`:
at minimum a `# <logical_name>` heading as the first
heading. Leaf nodes need frontmatter with `output`.

For ARTIFACT and external file tests, create the
referenced files on disk at the declared paths.

## Test cases

### Happy path

#### Simple leaf node — context and hash

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`
  heading, `# Public` with `## Context` subsection
  containing one line of content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`
  heading, frontmatter `output: out/a.txt`, `# Public`
  with `## Interface` subsection, `# Agent` section
  with content.
- Do not create `out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- First line matches `chain_hash: ` followed by
  exactly 27 non-whitespace characters.
- After `--- context ---`: contains `## Context`
  heading and its content, reduced frontmatter block
  with only `output: out/a.txt`, `## Interface` heading
  and its content, `# Agent` heading and agent content.
  No `# Public` headings appear.
- No `--- input ---` section.
- No `--- existing artifact ---` section.

#### Ancestor public content included

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Overview` with content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  `# Public` → `## Details` with content.
- Create `code-from-spec/a/b/_node.md` with
  `# SPEC/a/b`, frontmatter `output: out/b.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/a/b")`.

Expected:
- Context contains `## Overview` and `## Details`
  headings and their content. No `# Public` headings.

#### Ancestor without public section skipped

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`
  heading only (no public section).
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  `# Public` → `## Interface` with content,
  frontmatter `output: out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context does not contain SPEC's content.
- Contains `## Interface` heading and its content.

#### Ancestor with empty public section skipped

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` present but empty (no subsections).
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  `# Public` → `## Interface` with content,
  frontmatter `output: out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context does not contain SPEC's content.
- Contains `## Interface` heading and its content.

#### Dependency without qualifier — public included

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  `# Public` → `## Interface` + `## Constraints`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/b"]`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context contains `## Interface` and
  `## Constraints` headings and their content from
  SPEC/b.

#### Dependency with qualifier — subsection only

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  `# Public` → `## Interface` + `## Constraints`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/b(interface)"]`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context contains `## Interface` heading and content.
- Does not contain `## Constraints`.

#### ARTIFACT dependency — artifact tag line removed

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  frontmatter `output: out/b.go`.
- Create `out/b.go` with artifact tag line and body
  content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`,
  `depends_on: ["ARTIFACT/b"]`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context contains body content of `out/b.go`.
- Does not contain the artifact tag line.

#### EXTERNAL dependency — full content

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `data/config.yaml` with known content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `depends_on: ["EXTERNAL/data/config.yaml"]`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context contains the full content of
  `data/config.yaml`.

#### Target has reduced frontmatter with output only

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/b"]`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context contains a frontmatter block between `---`
  delimiters with only `output: out/a.txt`.
- Does not contain `depends_on`.

#### Target agent section included

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`, `# Public` →
  `## Interface` with content, `# Agent` with content.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Context contains `## Interface` and its content,
  `# Agent` and its content.
- No `# Public` heading.

#### Target without agent section — skipped

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`, `# Public` →
  `## Interface` with content. No `# Agent` section.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- No error. Context contains only public content.

#### Input present — in separate section

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/b/_node.md` with `# SPEC/b`,
  frontmatter `output: out/data.json`.
- Create `out/data.json` with artifact tag line and
  body content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `input: ARTIFACT/b`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- `--- input ---` section contains body of
  `out/data.json` without artifact tag line.
- Input content does not appear in context section.

#### EXTERNAL input — full content in input section

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `docs/vendor/spec.yaml` with known content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `input: EXTERNAL/docs/vendor/spec.yaml`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- `--- input ---` section contains the full content
  of `docs/vendor/spec.yaml`.

#### No input — section absent

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`. No input field.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Result does not contain `--- input ---`.

#### Existing artifact present — in separate section

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Create `out/a.go` with known content.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- `--- existing artifact ---` section contains the
  full content of `out/a.go`.

#### Existing artifact absent — section omitted

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.go`.
- Do not create `out/a.go`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Result does not contain `--- existing artifact ---`.

#### Hash is deterministic

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`,
  `# Public` → `## Overview` with stable content.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`.

Actions:
1. Call `MCPLoadChain("SPEC/a")` twice.

Expected:
- Both calls return identical `chain_hash` values.

### Error cases

#### Invalid logical name — not SPEC/

Actions:
1. Call `MCPLoadChain("INVALID/something")`.

Expected:
- Returns error `logicalnames.ErrUnsupportedReference`.

#### Nonexistent node file

Actions:
1. Call `MCPLoadChain("SPEC/nonexistent")` with no
   `_node.md` on disk.

Expected:
- Returns error propagated from `FrontmatterParse`
  (`file.ErrFileUnreadable`).

#### No output declared

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`.
  No output in frontmatter.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Returns error `ErrNoOutput`.

#### Invalid output path — traversal

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: ../../etc/passwd`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

Expected:
- Returns error `ErrInvalidOutputPath`.

#### Unresolvable dependency

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out/a.txt`,
  `depends_on: ["SPEC/missing"]`.
- Do not create `code-from-spec/missing/_node.md`.

Actions:
1. Call `MCPLoadChain("SPEC/a")`.

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
