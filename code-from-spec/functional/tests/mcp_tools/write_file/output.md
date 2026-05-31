<!-- code-from-spec: ROOT/functional/tests/mcp_tools/write_file@8Gyw58j9E8dn9yr1VUBzI819cBM -->

# Test Specification: MCPWriteFile

## Function under test

```
MCPWriteFile(logical_name, path, content) -> string
```

---

## Happy path

### Test: Writes file successfully

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "output/file.go"

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "output/file.go"
    content      = "package main"

Expected outcome:
  Return value equals "wrote output/file.go".
  A file exists on disk at <project_root>/output/file.go
  with content "package main".

---

### Test: Creates intermediate directories

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "deep/nested/dir/file.go"

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "deep/nested/dir/file.go"
    content      = "package main"

Expected outcome:
  Call succeeds (no error).
  All intermediate directories are created automatically.
  A file exists on disk at <project_root>/deep/nested/dir/file.go
  with content "package main".

---

### Test: Overwrites existing file

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "output/file.go"
  Create the file <project_root>/output/file.go on disk
  with content "old".

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "output/file.go"
    content      = "new"

Expected outcome:
  Call succeeds (no error).
  The file at <project_root>/output/file.go contains "new".
  The previous content "old" is gone.

---

## Error cases

### Test: Invalid logical name — ARTIFACT reference

Setup:
  No spec tree required.

Actions:
  Call MCPWriteFile with:
    logical_name = "ARTIFACT/x(y)"
    path         = "out.go"
    content      = ""

Expected outcome:
  Returns error UnsupportedReference.
  (Propagated from LogicalNames via LogicalNameToPath.)

---

### Test: Invalid logical name — with qualifier

Setup:
  No spec tree required.

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a(interface)"
    path         = "out.go"
    content      = ""

Expected outcome:
  Returns error UnsupportedReference.
  (LogicalNameToPath strips qualifiers; the resulting node
  path does not exist, so the frontmatter cannot be read.)

---

### Test: Nonexistent node file

Setup:
  No _node.md file exists for the path that corresponds
  to "ROOT/missing".

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/missing"
    path         = "out.go"
    content      = ""

Expected outcome:
  Returns error UnreadableFrontmatter.

---

### Test: No outputs declared

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter is empty (no outputs field).

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "out.go"
    content      = ""

Expected outcome:
  Returns error NoOutputs.

---

### Test: Path not in outputs

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "allowed/file.go"

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "other/file.go"
    content      = ""

Expected outcome:
  Returns error PathNotInOutputs.

---

### Test: Path validation — empty path

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "out.go"

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = ""
    content      = ""

Expected outcome:
  Returns error PathEmpty.
  (Propagated from PathUtils via PathValidateCfs.)

---

### Test: Path validation — directory traversal

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "out.go"

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "../../etc/passwd"
    content      = ""

Expected outcome:
  Returns error DirectoryTraversal.
  (Propagated from PathUtils via PathValidateCfs.)

---

### Test: Path validation — backslash in path

Setup:
  Create a spec tree with a node at ROOT/a.
  The node's _node.md frontmatter contains:
    outputs:
      - id: "code"
        path: "out.go"

Actions:
  Call MCPWriteFile with:
    logical_name = "ROOT/a"
    path         = "output\file.go"
    content      = ""

Expected outcome:
  Returns error PathContainsBackslash.
  (Propagated from PathUtils via PathValidateCfs.)
