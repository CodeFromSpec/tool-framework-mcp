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
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "modified content".
- Compute the current chain hash for SPEC/root/a.
- Create `.manifest` with entry for ARTIFACT/root/a
  with checksum that does NOT match the hash of
  "modified content" (simulating a modified file),
  and the current chain hash.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

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
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "artifact content".
- Compute the current checksum for "artifact content".
- Create `.manifest` with entry for ARTIFACT/root/a
  with the correct checksum but a stale chain hash
  (e.g. `AAAAAAAAAAAAAAAAAAAAAAAAAAA`).

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

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
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "new content".
- Create an empty `.manifest` (header only, no entries).
  Create the `.manifest.lock` file.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Return value = `"accepted out/a.go"`.
- Read manifest: entry for ARTIFACT/root/a exists
  with correct checksum and chain hash.

### Error cases

#### Not a SPEC reference

Actions:
1. Call `mcpaccept.MCPAccept("ARTIFACT/root/a")`.

Expected:
- Error `mcpaccept.ErrNotASpecReference`.

#### Nonexistent node file

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/missing")`.

Expected:
- Error `mcpaccept.ErrUnreadableFrontmatter`.

#### No output declared

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No output in frontmatter.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Error `mcpaccept.ErrNoOutput`.

#### Artifact file does not exist on disk

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Do not create `out/a.go` on disk.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Error propagated from oslayer (cannot read file
  to compute hash).

#### Already up to date

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "same content".
- Compute both the current checksum and chain hash.
- Create `.manifest` with entry for ARTIFACT/root/a
  with matching checksum and matching chain hash.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

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
