---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/mcp_tools/dump_chain
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcpdumpchain/mcpdumpchain_test.go
---

# SPEC/golang/test/cases/mcp_tools/dump_chain

# Agent

## Test setup guidance

`mcpdumpchain.MCPDumpChain` calls `mcploadchain.MCPLoadChain` internally and
writes the result to a file under `code-from-spec/.dump/`,
named after the logical name with slashes replaced by
underscores. For `SPEC/root/a`, the dump file is
`code-from-spec/.dump/SPEC_root_a.xml`. Tests must
create a valid spec tree on disk. Use `testutils.Chdir`
pattern.

## Test cases

### Happy path

#### Writes dump file

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`.

Actions:
1. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.

Expected:
- Return value = `"wrote code-from-spec/.dump/SPEC_root_a.xml"`.
- File `code-from-spec/.dump/SPEC_root_a.xml` exists on
  disk.
- Content starts with `<chain>`.
- Content contains `</chain>`.
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
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/root/a")`
   → `token`.
2. Call `mcploadchain.MCPLoadChain(token)` → store as
   `expected`.
3. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.
4. Read `code-from-spec/.dump/SPEC_root_a.xml` from disk.

Expected:
- File content equals `expected`.

#### Overwrites existing dump file

Setup:
- Create spec tree as above.
- Create `code-from-spec/.dump/SPEC_root_a.xml` with
  content "old".

Actions:
1. Call `mcpdumpchain.MCPDumpChain("SPEC/root/a")`.

Expected:
- `code-from-spec/.dump/SPEC_root_a.xml` contains the
  new chain, not "old".

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
- `code-from-spec/.dump/SPEC_root_a.xml` does not exist.

#### Invalid logical name

Actions:
1. Call `mcpdumpchain.MCPDumpChain("INVALID/something")`.

Expected:
- Error propagated from mcploadchain.MCPLoadChain.

## Go-specific guidance

- The package name is `mcpdumpchain_test` (external
  test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- Read the dump file with `os.ReadFile` to verify
  content.
- Import the `subagenttoken` package to mint the token
  needed for the direct `mcploadchain.MCPLoadChain` call
  used to compute `expected`.
