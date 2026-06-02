<!-- code-from-spec: ROOT/functional/tests/mcp_tools/chain_hash@nCwCkeTnF8A1_a5ZWLXyzsNMyH4 -->

# Test Specification: MCPChainHash

## Test cases

All tests create a spec tree on disk with `_node.md` files, then call `MCPChainHash`.

---

### Happy path

#### Returns a 27-character hash

Setup:
- Create ROOT/_node.md with a public section.
- Create ROOT/a/_node.md as a leaf with an output field.

Action:
- Call MCPChainHash with logical_name = "ROOT/a".

Expected outcome:
- Result is a string of exactly 27 characters.

---

#### Hash is deterministic

Setup:
- Create ROOT/_node.md with a public section containing known content.
- Create ROOT/a/_node.md as a leaf with an output field and known content.

Action:
- Call MCPChainHash with logical_name = "ROOT/a".
- Call MCPChainHash again with the same logical_name = "ROOT/a".

Expected outcome:
- Both calls return the same string.

---

#### Hash matches load_chain hash

Setup:
- Create ROOT/_node.md with a public section.
- Create ROOT/a/_node.md as a leaf with an output field.

Action:
- Call MCPChainHash with logical_name = "ROOT/a" — capture result as hash_a.
- Call MCPLoadChain with logical_name = "ROOT/a" — capture result.chain_hash as hash_b.

Expected outcome:
- hash_a equals hash_b.

---

### Error cases

#### Invalid logical name — not ROOT/

Setup:
- None.

Action:
- Call MCPChainHash with logical_name = "INVALID/something".

Expected outcome:
- Error UnsupportedReference is raised (propagated from LogicalNameToPath).

---

#### Nonexistent node file

Setup:
- No _node.md file exists at ROOT/nonexistent/.

Action:
- Call MCPChainHash with logical_name = "ROOT/nonexistent".

Expected outcome:
- Error FileUnreadable is raised (propagated from FrontmatterParse via FileReader).

---

#### No output declared

Setup:
- Create ROOT/_node.md with a public section.
- Create ROOT/a/_node.md as a leaf with no output field.

Action:
- Call MCPChainHash with logical_name = "ROOT/a".

Expected outcome:
- Error NoOutput is raised.
