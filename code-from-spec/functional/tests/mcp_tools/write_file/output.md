<!-- code-from-spec: ROOT/functional/tests/mcp_tools/write_file@44TJN42_jm-gdLALK1p-zpCfAl4 -->

## Test cases for MCPWriteFile

### Happy path

#### Writes file successfully

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "output/file.go"

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "output/file.go",
content = "package main".

Expected: return value = "wrote output/file.go". The file exists on disk with
content "package main".

---

#### Creates intermediate directories

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "deep/nested/dir/file.go"

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "deep/nested/dir/file.go",
content = "package main".

Expected: success. All intermediate directories are created automatically. The file
exists on disk.

---

#### Overwrites existing file

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "output/file.go"
Create "output/file.go" on disk with content "old".

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "output/file.go",
content = "new".

Expected: success. The file content on disk is "new".

---

### Error cases

#### Invalid logical name — ARTIFACT reference

Setup: none.

Action: call MCPWriteFile with logical_name = "ARTIFACT/x", path = "out.go",
content = "".

Expected: error UnsupportedReference, propagated from LogicalNames via
LogicalNameToPath.

---

#### Invalid logical name — with qualifier

Setup: none.

Action: call MCPWriteFile with logical_name = "ROOT/a(interface)", path = "out.go",
content = "".

Expected: error UnsupportedReference, propagated from LogicalNames via
LogicalNameToPath. LogicalNameToPath strips qualifiers, so this resolves but
the node file does not exist.

---

#### Nonexistent node file

Setup: no `_node.md` file exists at the path corresponding to ROOT/missing.

Action: call MCPWriteFile with logical_name = "ROOT/missing", path = "out.go",
content = "".

Expected: error UnreadableFrontmatter.

---

#### No output declared

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has empty
frontmatter (no output field).

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "out.go",
content = "".

Expected: error NoOutput.

---

#### Path not in output

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "allowed/file.go"

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "other/file.go",
content = "".

Expected: error PathNotInOutput.

---

#### Path validation — empty path

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "out.go"

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "", content = "".

Expected: error PathEmpty, propagated from PathUtils via PathValidateCfs.

---

#### Path validation — traversal

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "out.go"

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "../../etc/passwd",
content = "".

Expected: error DirectoryTraversal, propagated from PathUtils via PathValidateCfs.

---

#### Path validation — backslash

Setup: create a spec tree with a node at ROOT/a. The node's `_node.md` has frontmatter:
  output = "out.go"

Action: call MCPWriteFile with logical_name = "ROOT/a", path = "output\\file.go",
content = "".

Expected: error PathContainsBackslash, propagated from PathUtils via
PathValidateCfs.
