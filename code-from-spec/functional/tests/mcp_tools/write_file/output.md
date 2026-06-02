<!-- code-from-spec: ROOT/functional/tests/mcp_tools/write_file@wiWtZT9PMPpHChOJ4aeWSv8V6s4 -->

# Test cases for MCPWriteFile

All tests create a spec tree on disk with `_node.md` files containing
frontmatter with output declarations, then call `MCPWriteFile`.

---

## Happy path

### Writes file successfully

Setup: create spec tree with ROOT/a having frontmatter output = "output/file.go".

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "output/file.go",
content = "package main".

Expect:
- return value = "wrote output/file.go"
- the file exists on disk with content "package main"

---

### Creates intermediate directories

Setup: create spec tree with ROOT/a having frontmatter output = "deep/nested/dir/file.go".

Action: call MCPWriteFile with logical_name = "ROOT/a",
path = "deep/nested/dir/file.go", content = "package main".

Expect:
- success (return value = "wrote deep/nested/dir/file.go")
- intermediate directories are created automatically
- the file exists on disk

---

### Overwrites existing file

Setup: create spec tree with ROOT/a having frontmatter output = "output/file.go".
Create "output/file.go" on disk with initial content "old".

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "output/file.go",
content = "new".

Expect:
- success
- file content on disk is "new"

---

## Error cases

### Invalid logical name — ARTIFACT reference

Action: call MCPWriteFile with logical_name = "ARTIFACT/x", path = "out.go",
content = "".

Expect: error UnsupportedReference (propagated from LogicalNames via LogicalNameToPath).

---

### Invalid logical name — with qualifier

Action: call MCPWriteFile with logical_name = "ROOT/a(interface)", path = "out.go",
content = "".

Expect: error UnsupportedReference (propagated from LogicalNames — LogicalNameToPath
strips qualifiers, so this resolves but the node file won't exist).

---

### Nonexistent node file

Action: call MCPWriteFile with logical_name = "ROOT/missing"
(no _node.md file on disk), path = "out.go", content = "".

Expect: error UnreadableFrontmatter.

---

### No output declared

Setup: create spec tree with ROOT/a having empty frontmatter (no output field).

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "out.go",
content = "".

Expect: error NoOutput.

---

### Path not in output

Setup: create spec tree with ROOT/a having frontmatter output = "allowed/file.go".

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "other/file.go",
content = "".

Expect: error PathNotInOutput.

---

### Path validation — empty path

Setup: create spec tree with ROOT/a having frontmatter output = "out.go".

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "", content = "".

Expect: error PathEmpty (propagated from PathUtils via PathValidateCfs).

---

### Path validation — traversal

Setup: create spec tree with ROOT/a having frontmatter output = "out.go".

Action: call MCPWriteFile with logical_name = "ROOT/a",
path = "../../etc/passwd", content = "".

Expect: error DirectoryTraversal (propagated from PathUtils via PathValidateCfs).

---

### Path validation — backslash

Setup: create spec tree with ROOT/a having frontmatter output = "out.go".

Action: call MCPWriteFile with logical_name = "ROOT/a",
path = "output\\file.go", content = "".

Expect: error PathContainsBackslash (propagated from PathUtils via PathValidateCfs).
