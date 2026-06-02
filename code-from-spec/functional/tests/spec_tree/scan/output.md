<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@V76yZwm8XUEpcSb12YliOcHzYIU -->

# Test Specification: SpecTreeScan

## Happy Path

### Root node only

Setup: Create `code-from-spec/_node.md`.

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"ROOT"`
- `file_path` = `code-from-spec/_node.md`

No error.

---

### Root and nested nodes

Setup: Create the following files:
- `code-from-spec/_node.md`
- `code-from-spec/a/_node.md`
- `code-from-spec/a/b/_node.md`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with three `SpecTreeNode` entries:
- `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`
- `logical_name` = `"ROOT/a"`, `file_path` = `code-from-spec/a/_node.md`
- `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`

No error.

---

### Ignores non-node files

Setup: Create the following files:
- `code-from-spec/_node.md`
- `code-from-spec/x/output.md`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with exactly one entry: `logical_name` = `"ROOT"`. The `output.md` file is not included. No error.

---

### Ignores directories without _node.md

Setup: Create `code-from-spec/_node.md` and an empty subdirectory `code-from-spec/x/y/` (no files inside).

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with exactly one entry: `logical_name` = `"ROOT"`. The empty subdirectory produces no entries. No error.

---

### Result is sorted by logical name

Setup: Create the following files in non-alphabetical order:
- `code-from-spec/z/_node.md`
- `code-from-spec/_node.md`
- `code-from-spec/a/b/_node.md`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list of three entries sorted alphabetically by logical name:
1. `logical_name` = `"ROOT"`, `file_path` = `code-from-spec/_node.md`
2. `logical_name` = `"ROOT/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`
3. `logical_name` = `"ROOT/z"`, `file_path` = `code-from-spec/z/_node.md`

No error.

---

## Failure Cases

### No code-from-spec directory

Setup: Do not create a `code-from-spec/` directory.

Action: Call `SpecTreeScan`.

Expected outcome: Error propagated from `ListFiles` (directory not found).

---

### Empty code-from-spec directory

Setup: Create a `code-from-spec/` directory with no files inside.

Action: Call `SpecTreeScan`.

Expected outcome: Error `NoNodesFound`.

---

### Only non-node files in code-from-spec

Setup: Create the following files:
- `code-from-spec/README.md`
- `code-from-spec/x/output.md`

(No `_node.md` files anywhere.)

Action: Call `SpecTreeScan`.

Expected outcome: Error `NoNodesFound`.
