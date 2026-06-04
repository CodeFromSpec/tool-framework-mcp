<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@QTOlLUf16cGjVZJCMyfnieheGiw -->

## Test cases for SpecTreeScan

---

### Happy path

#### Root node only

Setup: Create `code-from-spec/_node.md`.

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with one `SpecTreeNode`:
- `logical_name` = `"ROOT"`
- `file_path` = `"code-from-spec/_node.md"`

---

#### Root and nested nodes

Setup: Create the following files:
- `code-from-spec/_node.md`
- `code-from-spec/a/_node.md`
- `code-from-spec/a/b/_node.md`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with three `SpecTreeNode` entries:
- `logical_name` = `"ROOT"`, `file_path` = `"code-from-spec/_node.md"`
- `logical_name` = `"ROOT/a"`, `file_path` = `"code-from-spec/a/_node.md"`
- `logical_name` = `"ROOT/a/b"`, `file_path` = `"code-from-spec/a/b/_node.md"`

---

#### Ignores non-node files

Setup: Create the following files:
- `code-from-spec/_node.md`
- `code-from-spec/x/output.md`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with one `SpecTreeNode`:
- `logical_name` = `"ROOT"`, `file_path` = `"code-from-spec/_node.md"`

The file `code-from-spec/x/output.md` is not included because it is not a `_node.md` file.

---

#### Ignores directories without _node.md

Setup: Create the following:
- `code-from-spec/_node.md`
- Empty subdirectory `code-from-spec/x/y/`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with one `SpecTreeNode`:
- `logical_name` = `"ROOT"`, `file_path` = `"code-from-spec/_node.md"`

The empty subdirectory is not included.

---

#### Result is sorted by logical name

Setup: Create the following files in non-alphabetical order:
- `code-from-spec/z/_node.md`
- `code-from-spec/_node.md`
- `code-from-spec/a/b/_node.md`

Action: Call `SpecTreeScan`.

Expected outcome: Returns a list with three `SpecTreeNode` entries sorted alphabetically by `logical_name`:
- `logical_name` = `"ROOT"`, `file_path` = `"code-from-spec/_node.md"`
- `logical_name` = `"ROOT/a/b"`, `file_path` = `"code-from-spec/a/b/_node.md"`
- `logical_name` = `"ROOT/z"`, `file_path` = `"code-from-spec/z/_node.md"`

---

### Failure cases

#### No code-from-spec directory

Setup: Do not create a `code-from-spec/` directory.

Action: Call `SpecTreeScan`.

Expected outcome: Error propagated from `ListFiles` indicating the directory was not found.

---

#### Empty code-from-spec directory

Setup: Create `code-from-spec/` with no files inside.

Action: Call `SpecTreeScan`.

Expected outcome: Error `NoNodesFound`.

---

#### Only non-node files in code-from-spec

Setup: Create the following files, but no `_node.md` files:
- `code-from-spec/README.md`
- `code-from-spec/x/output.md`

Action: Call `SpecTreeScan`.

Expected outcome: Error `NoNodesFound`.
