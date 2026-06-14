<!-- code-from-spec: ROOT/functional/tests/mcp_tools/write_file@bnQNzhf2ddB-FPg6Rb-l1_aYdLU -->

## Test suite: MCPWriteFile

---

### TC-01: Writes file successfully

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "output/file.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "output/file.go", content = "package main")

Expected outcome:
- Return value equals "wrote output/file.go"
- File exists on disk at "output/file.go"
- File content equals "package main"

---

### TC-02: Creates intermediate directories

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "deep/nested/dir/file.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "deep/nested/dir/file.go", content = "package main")

Expected outcome:
- Return value equals "wrote deep/nested/dir/file.go"
- Intermediate directories were created automatically
- File exists on disk at "deep/nested/dir/file.go"

---

### TC-03: Overwrites existing file

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "output/file.go"
- Create "output/file.go" on disk with content "old"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "output/file.go", content = "new")

Expected outcome:
- Return value equals "wrote output/file.go"
- File content on disk equals "new"

---

### TC-04: Error — invalid logical name, ARTIFACT reference

Setup:
- No spec tree required

Actions:
- Call MCPWriteFile(logical_name = "ARTIFACT/x", path = "out.go", content = "")

Expected outcome:
- Error UnsupportedReference raised (propagated from LogicalNames via LogicalNameToPath)

---

### TC-05: Error — invalid logical name, with qualifier

Setup:
- No spec tree required

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a(interface)", path = "out.go", content = "")

Expected outcome:
- Error QualifierNotAllowed raised

---

### TC-06: Error — nonexistent node file

Setup:
- No _node.md file exists for "SPEC/missing"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/missing", path = "out.go", content = "")

Expected outcome:
- Error UnreadableFrontmatter raised

---

### TC-07: Error — no output declared

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: empty (no output field)

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "out.go", content = "")

Expected outcome:
- Error NoOutput raised

---

### TC-08: Error — path not in output

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "allowed/file.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "other/file.go", content = "")

Expected outcome:
- Error PathNotInOutput raised

---

### TC-09: Error — empty path

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "out.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "", content = "")

Expected outcome:
- Error PathEmpty raised (propagated from PathUtils via PathValidateCfs)

---

### TC-10: Error — directory traversal

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "out.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "../../etc/passwd", content = "")

Expected outcome:
- Error DirectoryTraversal raised (propagated from PathUtils via PathValidateCfs)

---

### TC-11: Error — backslash in path

Setup:
- Create spec tree with node "SPEC/a"
- _node.md frontmatter: output = "out.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "output\\file.go", content = "")

Expected outcome:
- Error PathContainsBackslash raised (propagated from PathUtils via PathValidateCfs)
