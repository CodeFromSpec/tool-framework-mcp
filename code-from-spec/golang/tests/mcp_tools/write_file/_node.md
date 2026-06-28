---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/mcp_tools/write_file
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
output: internal/mcpwritefile/mcpwritefile_test.go
---

# SPEC/golang/tests/mcp_tools/write_file

# Agent

## Test setup guidance

`MCPWriteFile` reads the node's frontmatter from disk
to validate the path against the declared output. Tests
must create `_node.md` files with frontmatter containing
an output declaration. Use `testChdir` and create the
spec tree structure (`code-from-spec/.../_node.md`).

## Test cases

### Happy path

#### Writes file successfully

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: output/file.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a", "output/file.go",
   "package main")`.

Expected:
- Return value = `"wrote output/file.go"`.
- File exists on disk at `output/file.go` with
  content `"package main"`.

#### Creates intermediate directories

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: deep/nested/dir/file.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a",
   "deep/nested/dir/file.go", "package main")`.

Expected:
- Success. All intermediate directories created.
- File exists on disk.

#### Overwrites existing file

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: output/file.go`.
- Create `output/file.go` with content `"old"`.

Actions:
1. Call `MCPWriteFile("SPEC/a", "output/file.go",
   "new")`.

Expected:
- Success. File content is `"new"`.

### Error cases

#### Invalid logical name — ARTIFACT reference

Actions:
1. Call `MCPWriteFile("ARTIFACT/x", "out.go", "")`.

Expected:
- Error `ErrNotASpecReference`.

#### Invalid logical name — with qualifier

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a(interface)", "out.go",
   "")`.

Expected:
- Error `ErrQualifierNotAllowed`.

#### Nonexistent node file

Actions:
1. Call `MCPWriteFile("SPEC/missing", "out.go", "")`.

Expected:
- Error `ErrUnreadableFrontmatter`.

#### No output declared

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`.
  Empty frontmatter (no output).

Actions:
1. Call `MCPWriteFile("SPEC/a", "out.go", "")`.

Expected:
- Error `ErrNoOutput`.

#### Path not in output

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: allowed/file.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a", "other/file.go", "")`.

Expected:
- Error `ErrPathNotInOutput`.

#### Path validation — empty path

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a", "", "")`.

Expected:
- Error `pathutils.ErrPathEmpty`.

#### Path validation — traversal

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a", "../../etc/passwd",
   "")`.

Expected:
- Error `pathutils.ErrDirectoryTraversal`.

#### Path validation — backslash

Setup:
- Create `code-from-spec/_node.md` with `# SPEC`.
- Create `code-from-spec/a/_node.md` with `# SPEC/a`,
  frontmatter `output: out.go`.

Actions:
1. Call `MCPWriteFile("SPEC/a", "output\\file.go",
   "")`.

Expected:
- Error `pathutils.ErrPathContainsBackslash`.

## Go-specific guidance

- The package name is `mcpwritefile_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
