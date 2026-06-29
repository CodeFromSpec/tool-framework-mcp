---
depends_on:
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/mcp_tools/accept
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpaccept/mcpaccept_test.go
---

# SPEC/golang/tests/mcp_tools/accept

# Agent

## Test setup guidance

`MCPAccept` reads frontmatter and the manifest, then
updates the manifest checksum. Tests must create spec
tree files, output files, and manifest entries on disk.
Use `testChdir` pattern.

## Test cases

### Happy path

#### Accepts modified artifact

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "modified content".
- Create `.manifest` with entry for ARTIFACT/root/a
  with checksum that does NOT match the hash of
  "modified content" (simulating a modified file),
  and some chain hash.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Return value = `"accepted out/a.go"`.
- Read manifest: entry for ARTIFACT/root/a has
  Checksum updated to match the hash of
  "modified content". ChainHash is unchanged.

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

#### No manifest entry — not modified

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content.
- No `.manifest` file (or manifest without entry for
  ARTIFACT/root/a).

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Error `mcpaccept.ErrNotModified`.

#### Artifact file does not exist on disk

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `.manifest` with entry for ARTIFACT/root/a
  with some checksum and chain hash.
- Do not create `out/a.go` on disk.

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Error propagated from oslayer (cannot read file
  to compute hash).

#### Checksum already matches — not modified

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.
- Create `out/a.go` with content "same content".
- Create `.manifest` with entry for ARTIFACT/root/a
  with checksum matching the hash of "same content".

Actions:
1. Call `mcpaccept.MCPAccept("SPEC/root/a")`.

Expected:
- Error `mcpaccept.ErrNotModified`.

## Go-specific guidance

- The package name is `mcpaccept_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- Create `.manifest` files using
  `manifest.OpenManifest(false)` + `m.Save()`, or by
  writing the file directly.
- To compute file checksums for setup, use SHA-1 of
  the content (after CRLF→LF normalization, with
  trailing LF), encoded as base64url (27 chars).
