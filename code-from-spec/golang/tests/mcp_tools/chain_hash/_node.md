---
depends_on:
  - SPEC/golang/implementation/mcp_tools/chain_hash(interface)
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/utils/logical_names
output: internal/mcpchainhash/mcpchainhash_test.go
---

# SPEC/golang/tests/mcp_tools/chain_hash

Test cases for the chain hash tool.

# Agent

## Test cases

All tests create a spec tree on disk with `_node.md`
files, then call `MCPChainHash`.

### Happy path

#### Returns a 27-character hash

Create a spec tree: SPEC (with `# Public` containing a
`## Context` subsection) and SPEC/a (leaf with output).
Call MCPChainHash with logicalName = "SPEC/a".

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
error `logicalnames.ErrUnsupportedReference` (propagated
from LogicalNameToPath).

#### Nonexistent node file

Call MCPChainHash with "SPEC/nonexistent" (no _node.md
on disk). Expect error `file.ErrFileUnreadable`
(propagated from FrontmatterParse).

#### No output declared

Create a spec tree: SPEC, SPEC/a (leaf without
output). Call MCPChainHash with "SPEC/a". Expect
error `mcpchainhash.ErrNoOutput`.

## Go-specific guidance

- The package name is `mcpchainhash_test` (external test
  package).
- The "hash matches load_chain hash" test imports
  `mcploadchain` to call `MCPLoadChain` and compare
  the chain_hash field.
- Use `testChdir` and `testWriteFile` helpers for
  creating spec trees on disk.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
  Never place content directly under `# Public`
  without a subsection heading — this is a format
  error.
