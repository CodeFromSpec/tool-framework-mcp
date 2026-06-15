<!-- code-from-spec: SPEC/functional/tests/spec_tree/scan@g14U1VNFAZxPJdbdnqdHjdPwyf8 -->

## Test Suite: SpecTreeScan

---

### TC-01: Root node only

Setup:
- Create `code-from-spec/_node.md`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with one `SpecTreeNode`
- Entry has `logical_name` = `"SPEC"` and `file_path` = `code-from-spec/_node.md`

---

### TC-02: Root and nested nodes

Setup:
- Create `code-from-spec/_node.md`
- Create `code-from-spec/a/_node.md`
- Create `code-from-spec/a/b/_node.md`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with three `SpecTreeNode` entries
- Entry 1: `logical_name` = `"SPEC"`, `file_path` = `code-from-spec/_node.md`
- Entry 2: `logical_name` = `"SPEC/a"`, `file_path` = `code-from-spec/a/_node.md`
- Entry 3: `logical_name` = `"SPEC/a/b"`, `file_path` = `code-from-spec/a/b/_node.md`

---

### TC-03: Ignores non-node files

Setup:
- Create `code-from-spec/_node.md`
- Create `code-from-spec/x/output.md`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with exactly one `SpecTreeNode`
- Entry has `logical_name` = `"SPEC"`
- No entry for `code-from-spec/x/output.md` or `"SPEC/x"`

---

### TC-04: Ignores _-prefixed directories directly under code-from-spec

Setup:
- Create `code-from-spec/_node.md`
- Create `code-from-spec/_rules/some/_node.md`
- Create `code-from-spec/_tools/_node.md`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with exactly one `SpecTreeNode`
- Entry has `logical_name` = `"SPEC"`
- No entries for `_rules/` or `_tools/` subtrees

---

### TC-05: _-prefixed directories deeper in the tree are not ignored

Setup:
- Create `code-from-spec/_node.md`
- Create `code-from-spec/a/_node.md`
- Create `code-from-spec/a/_internal/_node.md`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with three `SpecTreeNode` entries
- Entry 1: `logical_name` = `"SPEC"`
- Entry 2: `logical_name` = `"SPEC/a"`
- Entry 3: `logical_name` = `"SPEC/a/_internal"`

---

### TC-06: Ignores directories without _node.md

Setup:
- Create `code-from-spec/_node.md`
- Create an empty subdirectory `code-from-spec/x/y/`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with exactly one `SpecTreeNode`
- Entry has `logical_name` = `"SPEC"`
- No entries for `code-from-spec/x/` or `code-from-spec/x/y/`

---

### TC-07: Result is sorted alphabetically by logical name

Setup:
- Create `code-from-spec/z/_node.md`
- Create `code-from-spec/_node.md`
- Create `code-from-spec/a/b/_node.md`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Returns a list with three entries in this exact order:
  1. `logical_name` = `"SPEC"`
  2. `logical_name` = `"SPEC/a/b"`
  3. `logical_name` = `"SPEC/z"`

---

### TC-08: No code-from-spec directory

Setup:
- Do not create a `code-from-spec/` directory

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Raises an error propagated from `ListFiles` indicating the directory was not found

---

### TC-09: Empty code-from-spec directory

Setup:
- Create `code-from-spec/` with no files inside

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Raises error `NoNodesFound`

---

### TC-10: Only non-node files in code-from-spec

Setup:
- Create `code-from-spec/README.md`
- Create `code-from-spec/x/output.md`
- No `_node.md` files anywhere under `code-from-spec/`

Actions:
- Call `SpecTreeScan()`

Expected outcome:
- Raises error `NoNodesFound`
