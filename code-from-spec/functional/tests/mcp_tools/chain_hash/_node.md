---
depends_on:
  - SPEC/functional/logic/mcp_tools/chain_hash(interface)
output: code-from-spec/functional/tests/mcp_tools/chain_hash/output.md
---

# SPEC/functional/tests/mcp_tools/chain_hash

Test cases for the chain hash tool.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files, then call `MCPChainHash`.

### Happy path

#### Returns a 27-character hash

Create a spec tree: SPEC (with `# Public` containing a
`## Context` subsection) and SPEC/a (leaf with output).
Call MCPChainHash with logical_name = "SPEC/a".

Expect result is a 27-character string.

#### Hash is deterministic

Create a spec tree with known content. Call
MCPChainHash twice with the same logical name.
Expect both results are identical.

#### Hash matches load_chain hash

Create a spec tree: SPEC (with `# Public` containing a
`## Context` subsection) and SPEC/a (leaf with output).
Call MCPChainHash with "SPEC/a" and call MCPLoadChain
with "SPEC/a".

Expect the hash from MCPChainHash equals the
chain_hash from MCPLoadChain.

### Error cases

#### Invalid logical name — not SPEC/

Call MCPChainHash with "INVALID/something". Expect
error logicalnames.UnsupportedReference (propagated
from LogicalNameToPath).

#### Nonexistent node file

Call MCPChainHash with "SPEC/nonexistent" (no _node.md
on disk). Expect error filereader.FileUnreadable
(propagated from FrontmatterParse).

#### No output declared

Create a spec tree: SPEC, SPEC/a (leaf without
output). Call MCPChainHash with "SPEC/a". Expect
error NoOutput.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `MCPChainHash`.
- Each test case creates a spec tree on disk with
  `_node.md` files, then calls `MCPChainHash`.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
  Never place content directly under `# Public`
  without a subsection heading — this is a format
  error.
