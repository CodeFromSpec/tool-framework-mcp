<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@DUDAJKBt7krislaGiFBs3mBKuTo -->

# Test Specification: SpecTreeScan

## Function Under Test

`SpecTreeScan() -> list of SpecTreeNode`

Each `SpecTreeNode` has:
- `logical_name`: string
- `file_path`: pathutils.PathCfs

---

## Happy Path Test Cases

### TC-01: Root node only

**Setup:**
- Create `code-from-spec/_node.md`.

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly one `SpecTreeNode`.
- The entry has `logical_name` = `"ROOT"` and
  `file_path` = `code-from-spec/_node.md`.

---

### TC-02: Root and nested nodes

**Setup:**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly three `SpecTreeNode` entries.
- Entry 1: `logical_name` = `"ROOT"`,
  `file_path` = `code-from-spec/_node.md`.
- Entry 2: `logical_name` = `"ROOT/a"`,
  `file_path` = `code-from-spec/a/_node.md`.
- Entry 3: `logical_name` = `"ROOT/a/b"`,
  `file_path` = `code-from-spec/a/b/_node.md`.

---

### TC-03: Ignores non-node files

**Setup:**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/x/output.md` (not a `_node.md` file).

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly one `SpecTreeNode`.
- The only entry has `logical_name` = `"ROOT"` and
  `file_path` = `code-from-spec/_node.md`.
- No entry exists for anything under `code-from-spec/x/`.

---

### TC-04: Ignores directories without _node.md

**Setup:**
- Create `code-from-spec/_node.md`.
- Create an empty subdirectory `code-from-spec/x/y/`
  (no files inside).

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly one `SpecTreeNode`.
- The only entry has `logical_name` = `"ROOT"` and
  `file_path` = `code-from-spec/_node.md`.
- No entries exist for `code-from-spec/x/` or
  `code-from-spec/x/y/`.

---

### TC-05: Result is sorted by logical name

**Setup:**
- Create `code-from-spec/z/_node.md`.
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns a list with exactly three `SpecTreeNode` entries
  sorted alphabetically by `logical_name`.
- Entry 1: `logical_name` = `"ROOT"`.
- Entry 2: `logical_name` = `"ROOT/a/b"`.
- Entry 3: `logical_name` = `"ROOT/z"`.

---

## Failure Case Test Cases

### TC-06: No code-from-spec directory

**Setup:**
- Do not create a `code-from-spec/` directory.

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns an error propagated from `ListFiles`
  indicating the directory was not found.
- No list is returned.

---

### TC-07: Empty code-from-spec directory

**Setup:**
- Create `code-from-spec/` with no files or subdirectories
  inside.

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns error `NoNodesFound`.
- No list is returned.

---

### TC-08: Only non-node files in code-from-spec

**Setup:**
- Create `code-from-spec/README.md`.
- Create `code-from-spec/x/output.md`.
- Do not create any `_node.md` files.

**Action:**
- Call `SpecTreeScan`.

**Expected outcome:**
- Returns error `NoNodesFound`.
- No list is returned.
