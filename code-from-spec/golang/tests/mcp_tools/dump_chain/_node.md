---
depends_on:
  - SPEC/golang/implementation/mcp_tools/dump_chain
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpdumpchain/mcpdumpchain_test.go
---

# SPEC/golang/tests/mcp_tools/dump_chain

# Agent

## Test setup guidance

`mcpdumpchain.MCPDumpChain` calls `mcploadchain.MCPLoadChain` internally and
writes the result to `dump_chain.xml`. Tests must
create a valid spec tree on disk. Use `testChdir`
pattern.

## Test cases

### Happy path

#### Writes dump_chain.xml

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.

Actions:
1. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.

Expected:
- Return value = `"wrote dump_chain.xml"`.
- File `dump_chain.xml` exists on disk.
- Content starts with `chain_hash: ` followed by
  27 characters.
- Content contains `<chain>` and `</chain>`.
- Content contains `<constraints>` with
  `<entry name="SPEC/root">`.

#### Content matches mcploadchain.MCPLoadChain output

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`,
  `# Agent` with content.

Actions:
1. Call `mcploadchain.MCPLoadChain("SPEC/root/a")` → store as
   `expected`.
2. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.
3. Read `dump_chain.xml` from disk.

Expected:
- File content equals `expected`.

#### Overwrites existing dump_chain.xml

Setup:
- Create spec tree as above.
- Create `dump_chain.xml` with content "old".

Actions:
1. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.

Expected:
- `dump_chain.xml` contains the new chain, not "old".

### Error cases

#### No output declared

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`. No output in frontmatter.

Actions:
1. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.

Expected:
- Error propagated from mcploadchain.MCPLoadChain (mcploadchain.ErrNoOutput).
- `dump_chain.xml` does not exist.

#### Invalid logical name

Actions:
1. Call `mcpdumpchain.MCPDumpChain("INVALID/something")`.

Expected:
- Error propagated from mcploadchain.MCPLoadChain.

## Go-specific guidance

- The package name is `mcpdumpchain_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- Read `dump_chain.xml` with `os.ReadFile` to verify
  content.
