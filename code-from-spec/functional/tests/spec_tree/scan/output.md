<!-- code-from-spec: ROOT/functional/tests/spec_tree/scan@R5cab_YxDcA2P-uL-MUNvywILV4 -->

## Test suite: SpecTreeScan

---

### TC-01: Root node only

**Setup**
Create `code-from-spec/_node.md`.

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"SPEC"`
- `file_path` = `code-from-spec/_node.md`

---

### TC-02: Root and nested nodes

**Setup**
Create:
- `code-from-spec/_node.md`
- `code-from-spec/a/_node.md`
- `code-from-spec/a/b/_node.md`

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with exactly three `SpecTreeNode` entries:
- `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`
- `logical_name` = `"SPEC/a"`, `file_path` = `code-from-spec/a/_node.md`
- `logical_name` = `"SPEC/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`

---

### TC-03: Ignores non-node files

**Setup**
Create:
- `code-from-spec/_node.md`
- `code-from-spec/x/output.md`

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`

The file `code-from-spec/x/output.md` is not included.

---

### TC-04: Ignores `_`-prefixed directories directly under `code-from-spec/`

**Setup**
Create:
- `code-from-spec/_node.md`
- `code-from-spec/_rules/some/_node.md`
- `code-from-spec/_tools/_node.md`

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`

Nodes inside `_rules/` and `_tools/` are excluded because those directories are `_`-prefixed and directly under `code-from-spec/`.

---

### TC-05: `_`-prefixed directories deeper in the tree are NOT ignored

**Setup**
Create:
- `code-from-spec/_node.md`
- `code-from-spec/a/_node.md`
- `code-from-spec/a/_internal/_node.md`

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with exactly three `SpecTreeNode` entries:
- `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`
- `logical_name` = `"SPEC/a"`, `file_path` = `code-from-spec/a/_node.md`
- `logical_name` = `"SPEC/a/_internal"`, `file_path` = `code-from-spec/a/_internal/_node.md`

The `_` prefix exclusion rule only applies to directories that are direct children of `code-from-spec/`.

---

### TC-06: Ignores directories without `_node.md`

**Setup**
Create:
- `code-from-spec/_node.md`
- An empty subdirectory `code-from-spec/x/y/` (no files inside)

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with exactly one `SpecTreeNode`:
- `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`

---

### TC-07: Result is sorted by logical name

**Setup**
Create:
- `code-from-spec/z/_node.md`
- `code-from-spec/_node.md`
- `code-from-spec/a/b/_node.md`

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns a list with three `SpecTreeNode` entries in this order:
1. `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`
2. `logical_name` = `"SPEC/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`
3. `logical_name` = `"SPEC/z"`, `file_path` = `code-from-spec/z/_node.md`

---

### TC-08: No `code-from-spec/` directory

**Setup**
Do not create a `code-from-spec/` directory.

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns an error propagated from `ListFiles` indicating the directory was not found.

---

### TC-09: Empty `code-from-spec/` directory

**Setup**
Create the `code-from-spec/` directory with no files or subdirectories inside.

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns error `NoNodesFound`.

---

### TC-10: Only non-node files in `code-from-spec/`

**Setup**
Create:
- `code-from-spec/README.md`
- `code-from-spec/x/output.md`

No `_node.md` files are created anywhere.

**Action**
Call `SpecTreeScan`.

**Expected outcome**
Returns error `NoNodesFound`.
