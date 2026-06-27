<!-- code-from-spec: SPEC/functional/tests/mcp_tools/write_file@hAJhuFwHe6QmUw0JC_AKRcmGAu0 -->

## Test suite: MCPWriteFile


### TC-01: Writes file successfully

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "output/file.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "output/file.go", content = "package main")

Expected:
- Return value equals "wrote output/file.go"
- File exists on disk at "output/file.go"
- File content equals "package main"


### TC-02: Creates intermediate directories

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "deep/nested/dir/file.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "deep/nested/dir/file.go", content = "package main")

Expected:
- Return value indicates success
- All intermediate directories are created automatically
- File exists on disk at "deep/nested/dir/file.go"


### TC-03: Overwrites existing file

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "output/file.go"
- Create "output/file.go" on disk with content "old"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "output/file.go", content = "new")

Expected:
- Return value indicates success
- File content on disk equals "new"


### TC-04: Invalid logical name — ARTIFACT reference

Setup:
- No files required

Actions:
- Call MCPWriteFile(logical_name = "ARTIFACT/x", path = "out.go", content = "")

Expected:
- Error UnsupportedReference (propagated from LogicalNames via LogicalNameToPath)


### TC-05: Invalid logical name — with qualifier

Setup:
- No files required

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a(interface)", path = "out.go", content = "")

Expected:
- Error QualifierNotAllowed


### TC-06: Nonexistent node file

Setup:
- No `_node.md` file exists for the node "SPEC/missing"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/missing", path = "out.go", content = "")

Expected:
- Error UnreadableFrontmatter


### TC-07: No output declared

Setup:
- Create `SPEC/a/_node.md` with empty frontmatter (no output field)

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "out.go", content = "")

Expected:
- Error NoOutput


### TC-08: Path not in output

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "allowed/file.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "other/file.go", content = "")

Expected:
- Error PathNotInOutput


### TC-09: Path validation — empty path

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "out.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "", content = "")

Expected:
- Error PathEmpty (propagated from PathUtils via PathValidateCfs)


### TC-10: Path validation — traversal

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "out.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "../../etc/passwd", content = "")

Expected:
- Error DirectoryTraversal (propagated from PathUtils via PathValidateCfs)


### TC-11: Path validation — backslash

Setup:
- Create `SPEC/a/_node.md` with frontmatter: output = "out.go"

Actions:
- Call MCPWriteFile(logical_name = "SPEC/a", path = "output\\file.go", content = "")

Expected:
- Error PathContainsBackslash (propagated from PathUtils via PathValidateCfs)
