---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/mcp_tools/accept
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpaccept/mcpaccept_test.go
---

# SPEC/golang/test/cases/mcp_tools/accept

# Agent

## Test setup guidance

`MCPAccept` reads frontmatter, computes checksum and
chain hash, then updates the manifest. Tests must
create spec tree files, output files, and manifest
entries on disk. Use the `testutils.Chdir` pattern.

To produce valid manifest entries with matching chain
hashes, use `chainresolver.ChainResolve` and
`chainhash.ChainHashCompute`. To compute file
checksums, use SHA-1 of the content (after CRLF→LF
normalization, with trailing LF), encoded as base64url
(27 chars).

## Test cases

### Happy path

#### Accepts modified artifact

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with content "modified content".
- Compute the current chain hash for SPEC/root/a.
- Create `.manifest` with entry for ARTIFACT/root/a
  with checksum that does NOT match the hash of
  "modified content" (simulating a modified file),
  and the current chain hash.

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Return value = `"accepted out/a.go"`.
- Read manifest: entry for ARTIFACT/root/a has
  Checksum updated to match the hash of
  "modified content". ChainHash unchanged.

#### Accepts stale artifact

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with content "artifact content".
- Compute the current checksum for "artifact content".
- Create `.manifest` with entry for ARTIFACT/root/a
  with the correct checksum but a stale chain hash
  (e.g. `AAAAAAAAAAAAAAAAAAAAAAAAAAA`).

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Return value = `"accepted out/a.go"`.
- Read manifest: entry for ARTIFACT/root/a has
  ChainHash updated to the current chain hash.
  Checksum unchanged.

#### Creates entry when none exists

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with content "new content".
- Create an empty `.manifest` (header only, no entries).
  Create the `.manifest.lock` file.

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Return value = `"accepted out/a.go"`.
- Read manifest: entry for ARTIFACT/root/a exists
  with correct checksum and chain hash.

#### Accepts verdict and sets result to accepted

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/v/_node.md` with
  `# SPEC/root/v`, frontmatter `type: verdict`,
  `output: code-from-spec/root/v/verdict.md`.
- Create `code-from-spec/root/v/verdict.md` with
  content "verdict content".
- Compute the current chain hash for SPEC/root/v.
- Create `.manifest` with entry for VERDICT/root/v
  with the correct checksum, chain hash, and
  `Result` = `"fail"`.

Actions:
1. Call `mcpaccept.MCPAccept("VERDICT/root/v")`.

Expected:
- Return value = `"accepted code-from-spec/root/v/verdict.md"`.
- Read manifest: entry for VERDICT/root/v has
  `Result` = `"accepted"`.

#### Accepts failed verdict even when hashes match

Setup:
- Create spec tree and verdict file as above.
- Compute both checksum and chain hash.
- Create `.manifest` with entry for VERDICT/root/v
  with matching checksum, matching chain hash, and
  `Result` = `"fail"`.

Actions:
1. Call `mcpaccept.MCPAccept("VERDICT/root/v")`.

Expected:
- Return value includes "accepted".
- Read manifest: `Result` = `"accepted"`.

### Error cases

#### Invalid prefix — SPEC reference

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Error `mcpaccept.ErrInvalidPrefix`.

#### Invalid prefix — EXTERNAL reference

Actions:
1. Call `mcpaccept.MCPAccept("EXTERNAL/file.txt")`.

Expected:
- Error `mcpaccept.ErrInvalidPrefix`.

#### Nonexistent node file

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/missing")`.

Expected:
- Error `mcpaccept.ErrUnreadableFrontmatter`.

#### No type declared

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No type in frontmatter.

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Error `mcpaccept.ErrNoOutput`.

#### File does not exist on disk

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Do not create `out/a.go` on disk.

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Error propagated from oslayer (cannot read file
  to compute hash).

#### Already up to date — artifact

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `type: artifact`, `output: out/a.go`.
- Create `out/a.go` with content "same content".
- Compute both the current checksum and chain hash.
- Create `.manifest` with entry for ARTIFACT/root/a
  with matching checksum and matching chain hash.

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Error `mcpaccept.ErrAlreadyUpToDate`.

#### Already up to date — verdict with result accepted

Setup:
- Create spec tree and verdict file.
- Compute checksum and chain hash.
- Create `.manifest` with VERDICT/ entry with matching
  checksum, matching chain hash, and
  `Result` = `"accepted"`.

Actions:
1. Call `mcpaccept.MCPAccept("VERDICT/root/v")`.

Expected:
- Error `mcpaccept.ErrAlreadyUpToDate`.

## Go-specific guidance

- The package name is `mcpaccept_test` (external test
  package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- Create `.manifest` files using
  `manifest.OpenManifest(false)` + `m.Save()`, or by
  writing the file directly.
- Use `chainresolver.ChainResolve` and
  `chainhash.ChainHashCompute` to compute valid chain
  hashes for test fixtures.
- To compute file checksums for setup, use SHA-1 of
  the content (after CRLF→LF normalization, with
  trailing LF), encoded as base64url (27 chars).
- Use `errors.Is` for error sentinel checks.
