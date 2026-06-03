<!-- code-from-spec: ROOT/functional/tests/mcp_tools/chain_hash@EaMoRn_pOxik8U0Ao7d4aSKenf8 -->

## Test: Returns a 27-character hash

Setup:
  Create a spec tree on disk:
    ROOT/_node.md — contains a public section
    ROOT/a/_node.md — leaf node with an output field

Actions:
  1. Call MCPChainHash with logical_name = "ROOT/a".

Expected outcome:
  Result is a string of exactly 27 characters.

---

## Test: Hash is deterministic

Setup:
  Create a spec tree on disk with known, fixed content:
    ROOT/_node.md — contains a public section
    ROOT/a/_node.md — leaf node with an output field

Actions:
  1. Call MCPChainHash with logical_name = "ROOT/a". Record result as first_hash.
  2. Call MCPChainHash with logical_name = "ROOT/a" again. Record result as second_hash.

Expected outcome:
  first_hash equals second_hash.

---

## Test: Hash matches load_chain hash

Setup:
  Create a spec tree on disk:
    ROOT/_node.md — contains a public section
    ROOT/a/_node.md — leaf node with an output field

Actions:
  1. Call MCPChainHash with logical_name = "ROOT/a". Record result as chain_hash_result.
  2. Call MCPLoadChain with logical_name = "ROOT/a". Record result as load_chain_result.
  3. Extract the chain_hash field from load_chain_result.

Expected outcome:
  chain_hash_result equals the chain_hash from load_chain_result.

---

## Test: Invalid logical name — not ROOT/

Setup:
  No spec tree required.

Actions:
  1. Call MCPChainHash with logical_name = "INVALID/something".

Expected outcome:
  Error logicalnames.UnsupportedReference is raised.

---

## Test: Nonexistent node file

Setup:
  Create a spec tree on disk:
    ROOT/_node.md — contains a public section
  Do not create ROOT/nonexistent/_node.md.

Actions:
  1. Call MCPChainHash with logical_name = "ROOT/nonexistent".

Expected outcome:
  Error filereader.FileUnreadable is raised.

---

## Test: No output declared

Setup:
  Create a spec tree on disk:
    ROOT/_node.md — contains a public section
    ROOT/a/_node.md — leaf node without an output field

Actions:
  1. Call MCPChainHash with logical_name = "ROOT/a".

Expected outcome:
  Error NoOutput is raised.
