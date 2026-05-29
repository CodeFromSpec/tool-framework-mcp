<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@sWncSTR5t30Djg205l8YAdimuvc -->

# Test Specification: SpecTreeScan

---

## Happy Path

---

### TC-01: Root node only

**Setup**
- Create `code-from-spec/_node.md`.

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.

---

### TC-02: Root and nested nodes

**Setup**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly three `SpecTreeNode` entries.
- Entry 1: `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`.
- Entry 2: `logical_name` = `"ROOT/a"`, `file_path` = `code-from-spec/a/_node.md`.
- Entry 3: `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`.

---

### TC-03: Ignores non-node files

**Setup**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/x/output.md` (not a `_node.md` file).

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.
- No entry is present for `code-from-spec/x/output.md`.

---

### TC-04: Ignores directories without _node.md

**Setup**
- Create `code-from-spec/_node.md`.
- Create an empty subdirectory `code-from-spec/x/y/` (no files inside).

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.
- No entries correspond to `code-from-spec/x/` or `code-from-spec/x/y/`.

---

### TC-05: Result is sorted by logical name

**Setup**
- Create `code-from-spec/z/_node.md`.
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly three `SpecTreeNode` entries in alphabetical order by `logical_name`:
  1. `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`.
  2. `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`.
  3. `logical_name` = `"ROOT/z"`, `file_path` = `code-from-spec/z/_node.md`.

---

## Failure Cases

---

### TC-06: No code-from-spec directory

**Setup**
- Do not create a `code-from-spec/` directory.

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- An error is returned, propagated from `ListFiles` indicating the directory was not found.
- No list of nodes is returned.

---

### TC-07: Empty code-from-spec directory

**Setup**
- Create `code-from-spec/` with no files inside.

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- An error `"no nodes found"` is returned.
- No list of nodes is returned.

---

### TC-08: Only non-node files in code-from-spec

**Setup**
- Create `code-from-spec/README.md`.
- Create `code-from-spec/x/output.md`.
- Do not create any `_node.md` files.

**Actions**
- Call `SpecTreeScan`.

**Expected outcome**
- An error `"no nodes found"` is returned.
- No list of nodes is returned.
