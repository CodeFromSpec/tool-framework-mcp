---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/mcp_tools/write_file
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpwritefile/mcpwritefile_test.go
---

# SPEC/golang/test/cases/mcp_tools/write_file

# Agent

## Test setup guidance

`MCPWriteFile` reads the node's frontmatter from disk
to derive the output path. Tests must create `_node.md`
files with frontmatter containing an output declaration.
Use `testutils.Chdir` and create the spec tree structure
(`code-from-spec/.../_node.md`).

## Test cases

### Happy path

#### Writes file successfully

Setup:
- Create `code-from-spec/root/_node.md` with `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with `# SPEC/root/a`,
  frontmatter `output: output/file.go`.

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/root/a",
   "package main")`.

Expected:
- Return value = `"wrote output/file.go"`.
- File exists on disk at `output/file.go` with
  content `"package main"`.

#### Manifest updated after write

Setup:
- Create `code-from-spec/root/_node.md` with `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with `# SPEC/root/a`,
  frontmatter `output: output/file.go`.

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/root/a",
   "package main")`.
2. Call `manifest.OpenManifest(true)`.

Expected:
- Manifest contains entry keyed `ARTIFACT/root/a`.
- Entry.Path = `output/file.go`.
- Entry.Checksum is a 27-character base64url string.
- Entry.ChainHash is a 27-character base64url string.

#### Creates intermediate directories

Setup:
- Create `code-from-spec/root/_node.md` with `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with `# SPEC/root/a`,
  frontmatter `output: deep/nested/dir/file.go`.

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/root/a",
   "package main")`.

Expected:
- Success. All intermediate directories created.
- File exists on disk.

#### Overwrites existing file

Setup:
- Create `code-from-spec/root/_node.md` with `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with `# SPEC/root/a`,
  frontmatter `output: output/file.go`.
- Create `output/file.go` with content `"old"`.

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/root/a", "new")`.

Expected:
- Success. File content is `"new"`.

### Error cases

#### Invalid logical name — ARTIFACT reference

Actions:
1. Call `mcpwritefile.MCPWriteFile("ARTIFACT/x", "")`.

Expected:
- Error `mcpwritefile.ErrNotASpecReference`.

#### Invalid logical name — with qualifier

Setup:
- Create `code-from-spec/root/_node.md` with `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with `# SPEC/root/a`,
  frontmatter `output: out.go`.

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/root/a(interface)",
   "")`.

Expected:
- Error `mcpwritefile.ErrQualifierNotAllowed`.

#### Nonexistent node file

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/missing", "")`.

Expected:
- Error `mcpwritefile.ErrUnreadableFrontmatter`.

#### No output declared

Setup:
- Create `code-from-spec/root/_node.md` with `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with `# SPEC/root/a`.
  Empty frontmatter (no output).

Actions:
1. Call `mcpwritefile.MCPWriteFile("SPEC/root/a", "")`.

Expected:
- Error `mcpwritefile.ErrNoOutput`.

## Go-specific guidance

- The package name is `mcpwritefile_test` (external
  test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
