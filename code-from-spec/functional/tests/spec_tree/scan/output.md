<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@sWncSTR5t30Djg205l8YAdimuvc -->

# Test Specification: SpecTreeScan

## Function Under Test

`SpecTreeScan() -> list of SpecTreeNode`

---

## Happy Path Cases

### Test 1: Root node only

**Setup:**
- Create `code-from-spec/_node.md`.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.

---

### Test 2: Root and nested nodes

**Setup:**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly three `SpecTreeNode` entries.
- Entry 1: `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`.
- Entry 2: `logical_name` = `"ROOT/a"`, `file_path` = `code-from-spec/a/_node.md`.
- Entry 3: `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`.

---

### Test 3: Ignores non-node files

**Setup:**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/x/output.md`.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.
- The file `code-from-spec/x/output.md` is not represented in the result.

---

### Test 4: Ignores directories without _node.md

**Setup:**
- Create `code-from-spec/_node.md`.
- Create an empty subdirectory `code-from-spec/x/y/`.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.
- The empty directory `code-from-spec/x/y/` does not produce any entries.

---

### Test 5: Result is sorted by logical name

**Setup:**
- Create `code-from-spec/z/_node.md`.
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly three `SpecTreeNode` entries.
- The list is sorted alphabetically by `logical_name`:
  1. `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`.
  2. `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`.
  3. `logical_name` = `"ROOT/z"`, `file_path` = `code-from-spec/z/_node.md`.

---

## Failure Cases

### Test 6: No code-from-spec directory

**Setup:**
- Do not create a `code-from-spec/` directory.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- An error is returned, propagated from `ListFiles`, indicating the directory was not found.
- No list is returned.

---

### Test 7: Empty code-from-spec directory

**Setup:**
- Create `code-from-spec/` with no files inside.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- An error `"no nodes found"` is returned.
- No list is returned.

---

### Test 8: Only non-node files in code-from-spec

**Setup:**
- Create `code-from-spec/README.md`.
- Create `code-from-spec/x/output.md`.
- Do not create any `_node.md` files.

**Actions:**
- Call `SpecTreeScan`.

**Expected outcome:**
- An error `"no nodes found"` is returned.
- No list is returned.
