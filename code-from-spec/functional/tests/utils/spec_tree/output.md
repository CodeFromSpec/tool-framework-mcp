<!-- code-from-spec: ROOT/functional/tests/utils/spec_tree@sWncSTR5t30Djg205l8YAdimuvc -->

# Test Specification: SpecTreeScan

---

## Happy Path

---

### Test: Root node only

**Setup**

Create file `code-from-spec/_node.md`.

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"ROOT"`
- `file_path` = `code-from-spec/_node.md`

---

### Test: Root and nested nodes

**Setup**

Create files:
- `code-from-spec/_node.md`
- `code-from-spec/a/_node.md`
- `code-from-spec/a/b/_node.md`

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns a list with exactly three `SpecTreeNode` entries:
- `logical_name` = `"ROOT"`,    `file_path` = `code-from-spec/_node.md`
- `logical_name` = `"ROOT/a"`,  `file_path` = `code-from-spec/a/_node.md`
- `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`

---

### Test: Ignores non-node files

**Setup**

Create files:
- `code-from-spec/_node.md`
- `code-from-spec/x/output.md`

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"ROOT"`
- `file_path` = `code-from-spec/_node.md`

The file `code-from-spec/x/output.md` is not included because it is not a `_node.md` file.

---

### Test: Ignores directories without _node.md

**Setup**

Create file `code-from-spec/_node.md`.
Create empty subdirectory `code-from-spec/x/y/` (no files inside).

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"ROOT"`
- `file_path` = `code-from-spec/_node.md`

The empty subdirectory does not produce any entries.

---

### Test: Result is sorted by logical name

**Setup**

Create files in non-alphabetical order:
- `code-from-spec/z/_node.md`
- `code-from-spec/_node.md`
- `code-from-spec/a/b/_node.md`

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns a list with exactly three `SpecTreeNode` entries in alphabetical order by `logical_name`:
1. `logical_name` = `"ROOT"`,    `file_path` = `code-from-spec/_node.md`
2. `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`
3. `logical_name` = `"ROOT/z"`,  `file_path` = `code-from-spec/z/_node.md`

---

## Failure Cases

---

### Test: No code-from-spec directory

**Setup**

Do not create a `code-from-spec/` directory.

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns an error propagated from `ListFiles` indicating the directory was not found.
No list is returned.

---

### Test: Empty code-from-spec directory

**Setup**

Create directory `code-from-spec/` with no files inside.

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns error `"no nodes found"`.
No list is returned.

---

### Test: Only non-node files in code-from-spec

**Setup**

Create files:
- `code-from-spec/README.md`
- `code-from-spec/x/output.md`

Do not create any `_node.md` files.

**Action**

Call `SpecTreeScan`.

**Expected outcome**

Returns error `"no nodes found"`.
No list is returned.
