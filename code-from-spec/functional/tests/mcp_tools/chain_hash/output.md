<!-- code-from-spec: ROOT/functional/tests/mcp_tools/chain_hash@0O8kf-roEzZ7-P94-oyYGqaqsrw -->

## Test suite: MCPChainHash

---

### Happy path

---

#### Returns a 27-character hash

Setup:
- Create a spec tree on disk:
  - `ROOT/_node.md` with a public section.
  - `ROOT/a/_node.md` as a leaf node with an output field.

Actions:
1. Call `MCPChainHash` with `logical_name` = `"ROOT/a"`.

Expected outcome:
- The result is a string with exactly 27 characters.

---

#### Hash is deterministic

Setup:
- Create a spec tree on disk with known, fixed content:
  - `ROOT/_node.md` with a public section.
  - `ROOT/a/_node.md` as a leaf node with an output field.

Actions:
1. Call `MCPChainHash` with `logical_name` = `"ROOT/a"`. Record the result as `hash1`.
2. Call `MCPChainHash` again with `logical_name` = `"ROOT/a"`. Record the result as `hash2`.

Expected outcome:
- `hash1` equals `hash2`.

---

#### Hash matches load_chain hash

Setup:
- Create a spec tree on disk:
  - `ROOT/_node.md` with a public section.
  - `ROOT/a/_node.md` as a leaf node with an output field.

Actions:
1. Call `MCPChainHash` with `logical_name` = `"ROOT/a"`. Record the result as `chain_hash_result`.
2. Call `MCPLoadChain` with `logical_name` = `"ROOT/a"`. Extract the `chain_hash` value from the returned document's first line.

Expected outcome:
- `chain_hash_result` equals the `chain_hash` extracted from `MCPLoadChain`.

---

### Error cases

---

#### Invalid logical name — not ROOT/

Setup:
- No special disk setup required.

Actions:
1. Call `MCPChainHash` with `logical_name` = `"INVALID/something"`.

Expected outcome:
- An error is returned.
- The error is `UnsupportedReference`, propagated from `LogicalNameToPath`.

---

#### Nonexistent node file

Setup:
- Create a spec tree on disk:
  - `ROOT/_node.md` with a public section.
  - No `ROOT/nonexistent/_node.md` file exists.

Actions:
1. Call `MCPChainHash` with `logical_name` = `"ROOT/nonexistent"`.

Expected outcome:
- An error is returned.
- The error is `FileUnreadable`, propagated from `FrontmatterParse` via `FileReader`.

---

#### No output declared

Setup:
- Create a spec tree on disk:
  - `ROOT/_node.md` with a public section.
  - `ROOT/a/_node.md` as a leaf node without an output field.

Actions:
1. Call `MCPChainHash` with `logical_name` = `"ROOT/a"`.

Expected outcome:
- An error is returned.
- The error is `NoOutput`.
