---
depends_on:
  - ROOT/functional/logic/mcp_tools/write_file(interface)
output: code-from-spec/functional/tests/mcp_tools/write_file/output.md
---

# ROOT/functional/tests/mcp_tools/write_file

Test cases for the write file tool.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files containing frontmatter with output declarations,
then call `MCPWriteFile`.

### Happy path

#### Writes file successfully

Create a spec tree with ROOT/a having output = "output/file.go". Call
MCPWriteFile with logical_name = "ROOT/a", path =
"output/file.go", content = "package main".

Expect return value = "wrote output/file.go". Verify
the file exists on disk with content "package main".

#### Creates intermediate directories

Create a spec tree with ROOT/a having output = "deep/nested/dir/file.go". Call
MCPWriteFile with path = "deep/nested/dir/file.go",
content = "package main".

Expect success. Directories created automatically.
Verify the file exists.

#### Overwrites existing file

Create a spec tree with ROOT/a having output = "output/file.go". Create
"output/file.go" on disk with initial content "old".
Call MCPWriteFile with path = "output/file.go",
content = "new".

Expect success. Verify file content is "new".

### Error cases

#### Invalid logical name — ARTIFACT reference

Call MCPWriteFile with logical_name = "ARTIFACT/x",
path = "out.go", content = "". Expect error UnsupportedReference (propagated from
LogicalNames via LogicalNameToPath).

#### Invalid logical name — with qualifier

Call MCPWriteFile with logical_name =
"ROOT/a(interface)", path = "out.go", content = "".
Expect error UnsupportedReference (propagated from
LogicalNames — LogicalNameToPath strips qualifiers,
so this resolves but the node file won't exist).

#### Nonexistent node file

Call MCPWriteFile with logical_name = "ROOT/missing"
(no _node.md file on disk), path = "out.go",
content = "". Expect error UnreadableFrontmatter.

#### No output declared

Create a spec tree with ROOT/a having empty frontmatter
(no output). Call MCPWriteFile with logical_name =
"ROOT/a", path = "out.go", content = "".

Expect error NoOutput.

#### Path not in output

Create a spec tree with ROOT/a having output =
"allowed/file.go". Call
MCPWriteFile with logical_name = "ROOT/a", path =
"other/file.go", content = "".

Expect error PathNotInOutput.

#### Path validation — empty path

Create a spec tree with ROOT/a having output =
"out.go". Call MCPWriteFile
with logical_name = "ROOT/a", path = "", content = "".

Expect error PathEmpty (propagated from PathUtils
via PathValidateCfs).

#### Path validation — traversal

Create a spec tree with ROOT/a having output =
"out.go". Call MCPWriteFile
with logical_name = "ROOT/a", path =
"../../etc/passwd", content = "".

Expect error DirectoryTraversal (propagated from
PathUtils via PathValidateCfs).

#### Path validation — backslash

Create a spec tree with ROOT/a having output =
"out.go". Call MCPWriteFile
with logical_name = "ROOT/a", path =
"output\\file.go", content = "".

Expect error PathContainsBackslash (propagated from
PathUtils via PathValidateCfs).

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `MCPWriteFile`.
- Each test case creates a spec tree on disk with
  `_node.md` files, then calls `MCPWriteFile`.
- Describe setup as files to create with their
  frontmatter content.
