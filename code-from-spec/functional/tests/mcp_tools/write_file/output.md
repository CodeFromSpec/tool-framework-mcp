<!-- code-from-spec: ROOT/functional/tests/mcp_tools/write_file@e4Zm-W2SxHLTBjuGJS2NNaekQMg -->

# Test Specification: MCPWriteFile

## Function Under Test

```
MCPWriteFile(logical_name: string, path: string, content: string) -> string
```

---

## Happy Path

### TC-01: Writes file successfully

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "output/file.go"
```

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"output/file.go"`
- `content` = `"package main"`

**Expected outcome**

- Return value equals `"wrote output/file.go"`.
- File exists on disk at `<cfs_root>/output/file.go`.
- File content equals `"package main"`.

---

### TC-02: Creates intermediate directories

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "deep/nested/dir/file.go"
```

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"deep/nested/dir/file.go"`
- `content` = `"package main"`

**Expected outcome**

- Return value equals `"wrote deep/nested/dir/file.go"`.
- All intermediate directories are created automatically.
- File exists on disk at `<cfs_root>/deep/nested/dir/file.go`.
- File content equals `"package main"`.

---

### TC-03: Overwrites existing file

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "output/file.go"
```

Create the file `<cfs_root>/output/file.go` on disk with content `"old"`.

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"output/file.go"`
- `content` = `"new"`

**Expected outcome**

- Return value equals `"wrote output/file.go"`.
- File content on disk equals `"new"`.

---

## Error Cases

### TC-04: Invalid logical name — ARTIFACT reference

**Setup**

No spec tree required.

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ARTIFACT/x(y)"`
- `path` = `"out.go"`
- `content` = `""`

**Expected outcome**

- Error `UnsupportedReference` is returned.
- Propagated from `LogicalNames` via `LogicalNameToPath`.

---

### TC-05: Invalid logical name — with qualifier

**Setup**

No spec tree required.

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a(interface)"`
- `path` = `"out.go"`
- `content` = `""`

**Expected outcome**

- Error `UnsupportedReference` is returned.
- Propagated from `LogicalNames` via `LogicalNameToPath`.
  (Qualifiers are stripped before path resolution; the resulting node file does not exist.)

---

### TC-06: Nonexistent node file

**Setup**

No `_node.md` file exists at the path corresponding to `ROOT/missing`.

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/missing"`
- `path` = `"out.go"`
- `content` = `""`

**Expected outcome**

- Error `UnreadableFrontmatter` is returned.

---

### TC-07: No outputs declared

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter is empty (no `outputs` field).

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"out.go"`
- `content` = `""`

**Expected outcome**

- Error `NoOutputs` is returned.

---

### TC-08: Path not in outputs

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "allowed/file.go"
```

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"other/file.go"`
- `content` = `""`

**Expected outcome**

- Error `PathNotInOutputs` is returned.

---

### TC-09: Path validation — empty path

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "out.go"
```

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `""`
- `content` = `""`

**Expected outcome**

- Error `PathEmpty` is returned.
- Propagated from `PathUtils` via `PathValidateCfs`.

---

### TC-10: Path validation — directory traversal

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "out.go"
```

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"../../etc/passwd"`
- `content` = `""`

**Expected outcome**

- Error `DirectoryTraversal` is returned.
- Propagated from `PathUtils` via `PathValidateCfs`.

---

### TC-11: Path validation — backslash

**Setup**

Create a spec tree on disk with a `_node.md` file at node `ROOT/a`.
The file's frontmatter contains:

```
outputs:
  - id: "code"
    path: "out.go"
```

**Action**

Call `MCPWriteFile` with:
- `logical_name` = `"ROOT/a"`
- `path` = `"output\\file.go"`
- `content` = `""`

**Expected outcome**

- Error `PathContainsBackslash` is returned.
- Propagated from `PathUtils` via `PathValidateCfs`.
