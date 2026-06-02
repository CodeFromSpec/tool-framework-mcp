---
depends_on:
  - ROOT/functional/logic/mcp_tools/chain_hash(interface)
outputs:
  - id: chain_hash_tests
    path: code-from-spec/functional/tests/mcp_tools/chain_hash/output.md
---

# ROOT/functional/tests/mcp_tools/chain_hash

Test cases for the chain hash tool.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files, then call `MCPChainHash`.

### Happy path

#### Returns a 27-character hash

Create a spec tree: ROOT (with public section) and
ROOT/a (leaf with output). Call MCPChainHash with
logical_name = "ROOT/a".

Expect result is a 27-character string.

#### Hash is deterministic

Create a spec tree with known content. Call
MCPChainHash twice with the same logical name.
Expect both results are identical.

#### Hash matches load_chain hash

Create a spec tree: ROOT (with public section) and
ROOT/a (leaf with output). Call MCPChainHash with
"ROOT/a" and call MCPLoadChain with "ROOT/a".

Expect the hash from MCPChainHash equals the
chain_hash from MCPLoadChain.

### Error cases

#### Invalid logical name — not ROOT/

Call MCPChainHash with "INVALID/something". Expect
error UnsupportedReference (propagated from
LogicalNameToPath).

#### Nonexistent node file

Call MCPChainHash with "ROOT/nonexistent" (no _node.md
on disk). Expect error propagated from
FrontmatterParse (FileUnreadable).

#### No output declared

Create a spec tree: ROOT, ROOT/a (leaf without
output). Call MCPChainHash with "ROOT/a". Expect
error NoOutput.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `MCPChainHash`.
- Each test case creates a spec tree on disk with
  `_node.md` files, then calls `MCPChainHash`.
