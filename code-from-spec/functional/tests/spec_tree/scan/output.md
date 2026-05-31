<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@3oU5PLHwcGCjvoTa7lpI-4kW01Q -->

# Test Specification: SpecTreeScan

## Interface

```
function SpecTreeScan() -> list of SpecTreeNode
```

Each `SpecTreeNode` has:
- `logical_name`: string
- `file_path`: PathCfs

---

## Happy Path Tests

### TC-1: Root node only

**Setup**
- Create `code-from-spec/_node.md`.

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly one entry.
- Entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.

---

### TC-2: Root and nested nodes

**Setup**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly three entries.
- Entry 1: `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`.
- Entry 2: `logical_name` = `"ROOT/a"`, `file_path` = `code-from-spec/a/_node.md`.
- Entry 3: `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`.

---

### TC-3: Ignores non-node files

**Setup**
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/x/output.md` (not a `_node.md`).

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly one entry.
- Entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.
- No entry for `code-from-spec/x/output.md`.

---

### TC-4: Ignores directories without _node.md

**Setup**
- Create `code-from-spec/_node.md`.
- Create an empty subdirectory `code-from-spec/x/y/` (no files inside).

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly one entry.
- Entry has `logical_name` = `"ROOT"` and `file_path` = `code-from-spec/_node.md`.
- No entries corresponding to `code-from-spec/x/y/`.

---

### TC-5: Result is sorted by logical name

**Setup**
- Create `code-from-spec/z/_node.md`.
- Create `code-from-spec/_node.md`.
- Create `code-from-spec/a/b/_node.md`.

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns a list with exactly three entries in alphabetical order by `logical_name`:
  1. `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`.
  2. `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`.
  3. `logical_name` = `"ROOT/z"`, `file_path` = `code-from-spec/z/_node.md`.

---

## Failure Case Tests

### TC-6: No code-from-spec directory

**Setup**
- Do not create a `code-from-spec/` directory.

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns an error propagated from `ListFiles` (directory not found).
- Does not return a list of nodes.

---

### TC-7: Empty code-from-spec directory

**Setup**
- Create `code-from-spec/` directory with no files inside.

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns error `NoNodesFound`.
- Does not return a list of nodes.

---

### TC-8: Only non-node files in code-from-spec

**Setup**
- Create `code-from-spec/README.md`.
- Create `code-from-spec/x/output.md`.
- Do not create any `_node.md` files.

**Action**
- Call `SpecTreeScan`.

**Expected outcome**
- Returns error `NoNodesFound`.
- Does not return a list of nodes.
