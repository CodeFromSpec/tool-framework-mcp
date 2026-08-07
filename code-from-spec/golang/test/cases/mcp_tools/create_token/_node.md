---
depends_on:
  - SPEC/golang/implementation/mcp_tools/create_token
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcpcreatetoken/mcpcreatetoken_test.go
---

# SPEC/golang/test/cases/mcp_tools/create_token

# Agent

## Test setup guidance

This package is pure (no file I/O), so tests do not need
`testutils.Chdir`.

## Test cases

### Happy path

#### Returns a token that decodes back to the logical name

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/a/b")` → `token`.
2. Call `subagenttoken.SubagentTokenValidate(token)` →
   `logicalName`.

Expected: No errors from either call. `logicalName`
equals `"SPEC/a/b"`.

#### Two calls for the same logical name produce different tokens

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/a")` twice →
   `token1`, `token2`.

Expected: No errors. `token1` differs from `token2`.

### Error cases

#### ARTIFACT reference — invalid logical name

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("ARTIFACT/x")`.

Expected: Error `mcpcreatetoken.ErrNotASpecReference`.

#### EXTERNAL reference — invalid logical name

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("EXTERNAL/x")`.

Expected: Error `mcpcreatetoken.ErrNotASpecReference`.

#### Qualifier not allowed

Actions:
1. Call `mcpcreatetoken.MCPCreateToken("SPEC/a(interface)")`.

Expected: Error `mcpcreatetoken.ErrQualifierNotAllowed`.

## Go-specific guidance

- The package name is `mcpcreatetoken_test` (external
  test package).
- Use `errors.Is` to check error sentinels.
- Import the `subagenttoken` package to verify round
  trips via `SubagentTokenValidate`.
