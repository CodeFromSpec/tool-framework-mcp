---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/test/utils/helpers
  - SPEC/golang/implementation/mcp_tools/write_verdict
  - SPEC/golang/implementation/mcp_tools/create_token
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcpwriteverdict/mcpwriteverdict_test.go
---

# SPEC/golang/test/cases/mcp_tools/write_verdict

Tests for the `MCPWriteVerdict` tool.

# Agent

## Test setup guidance

Each test must call `testutils.Chdir(t)` to set up a
temp directory as the working directory.

To get a valid token: create a spec node with
`type: verdict` and the desired `output` (or use the
default), then call
`mcpcreatetoken.MCPCreateToken("SPEC/<path>")`.

## Test cases

### Happy path

#### Write verdict with pass result

Setup:
- Create spec node `SPEC/root` with `# SPEC/root`.
- Create spec node `SPEC/root/v` with
  `type: verdict`,
  `output: code-from-spec/root/v/verdict.md`.

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/root/v")` → token.
2. Call `mcpwriteverdict.MCPWriteVerdict(token, true, "All checks passed.\n")`.

Expected:
- Returns `"wrote code-from-spec/root/v/verdict.md"`.
- File exists at `code-from-spec/root/v/verdict.md`
  with the provided content.
- Manifest contains a `VERDICT/root/v` entry with
  `Result` = `"pass"`, non-empty Checksum and ChainHash.

#### Write verdict with fail result

Setup:
- Create spec node `SPEC/root` with `# SPEC/root`.
- Create spec node `SPEC/root/v` with
  `type: verdict`,
  `output: code-from-spec/root/v/verdict.md`.

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/root/v")` → token.
2. Call `mcpwriteverdict.MCPWriteVerdict(token, false, "Found issues.\n")`.

Expected:
- Returns `"wrote code-from-spec/root/v/verdict.md"`.
- Manifest contains `VERDICT/root/v` entry with
  `Result` = `"fail"`.

#### Write verdict with default output path

Setup:
- Create spec node `SPEC/root` with `# SPEC/root`.
- Create spec node `SPEC/root/v` with
  `type: verdict` and no `output` field.

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/root/v")` → token.
2. Call `mcpwriteverdict.MCPWriteVerdict(token, true, "OK.\n")`.

Expected:
- Returns `"wrote code-from-spec/root/v/verdict.md"`.
- File exists at `code-from-spec/root/v/verdict.md`.

### Error cases

#### Node has no type — ErrNoOutput

Setup:
- Create spec node `SPEC/root` with `# SPEC/root`.
- Create spec node `SPEC/root/a` with no type field.

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/root/a")` → token.
2. Call `mcpwriteverdict.MCPWriteVerdict(token, true, "content")`.

Expected: Error `mcpwriteverdict.ErrNoOutput`.

#### Node has type artifact — ErrNotAVerdict

Setup:
- Create spec node `SPEC/root` with `# SPEC/root`.
- Create spec node `SPEC/root/a` with `type: artifact`,
  `output: internal/x.go`.

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/root/a")` → token.
2. Call `mcpwriteverdict.MCPWriteVerdict(token, true, "content")`.

Expected: Error `mcpwriteverdict.ErrNotAVerdict`.

#### Invalid token

Actions:
1. Call `mcpwriteverdict.MCPWriteVerdict("bad-token", true, "content")`.

Expected: Error from subagenttoken validation.

## Go-specific guidance

- The package name is `mcpwriteverdict_test` (external
  test package).
- Use `testutils.Chdir(t)` for each test.
- Use `testutils.CreateSpecNode` for node setup.
- Use `errors.Is` for sentinel error checks.
- Read the manifest via `manifest.OpenManifest(true)` to
  verify entries after write.
