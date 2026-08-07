---
depends_on:
  - SPEC/golang/implementation/subagent_token(import)
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/subagenttoken/subagenttoken_test.go
---

# SPEC/golang/test/cases/subagent_token

# Agent

## Test setup guidance

This package is pure (no file I/O), so tests do not need
`testutils.Chdir`.

## Test cases

### Round trip

#### Validate returns the original logical name

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/a/b")` → `token`.
2. Call `subagenttoken.SubagentTokenValidate(token)` → `logicalName`.

Expected: No errors. `logicalName` equals `"SPEC/a/b"`.

#### Round trip with an ARTIFACT logical name

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("ARTIFACT/x/y")` → `token`.
2. Call `subagenttoken.SubagentTokenValidate(token)` → `logicalName`.

Expected: No errors. `logicalName` equals `"ARTIFACT/x/y"`.

#### Round trip with an empty logical name

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("")` → `token`.
2. Call `subagenttoken.SubagentTokenValidate(token)` → `logicalName`.

Expected: No errors. `logicalName` equals `""`.

### Token properties

#### Tokens for the same logical name differ across calls

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/a")` twice →
   `token1`, `token2`.

Expected: No errors. `token1` differs from `token2` (random
nonce per call).

#### Tokens for different logical names differ

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/a")` → `token1`.
2. Call `subagenttoken.SubagentTokenGenerate("SPEC/b")` → `token2`.

Expected: No errors. `token1` differs from `token2`.

#### Token is URL-safe base64 text

Actions:
1. Call `subagenttoken.SubagentTokenGenerate("SPEC/a")` → `token`.

Expected: No error. `token` contains only characters from
`A-Z`, `a-z`, `0-9`, `-`, `_`.

### Error cases

#### Malformed base64 — invalid token

Actions:
1. Call `subagenttoken.SubagentTokenValidate("not-valid-base64!!!")`.

Expected: Error `subagenttoken.ErrInvalidToken`.

#### Too short to contain a nonce — invalid token

Actions:
1. Call `subagenttoken.SubagentTokenValidate("QQ")` (decodes to
   fewer bytes than the GCM nonce size).

Expected: Error `subagenttoken.ErrInvalidToken`.

#### Tampered ciphertext fails authentication

Setup:
- Call `subagenttoken.SubagentTokenGenerate("SPEC/a")` → `token`.
- Decode `token` with `base64.RawURLEncoding`, flip one bit in
  the last byte, re-encode with `base64.RawURLEncoding` →
  `tamperedToken`.

Actions:
1. Call `subagenttoken.SubagentTokenValidate(tamperedToken)`.

Expected: Error `subagenttoken.ErrInvalidToken`.

#### Empty string — invalid token

Actions:
1. Call `subagenttoken.SubagentTokenValidate("")`.

Expected: Error `subagenttoken.ErrInvalidToken`.

## Go-specific guidance

- The package name is `subagenttoken_test` (external test
  package).
- Use `errors.Is` to check `subagenttoken.ErrInvalidToken`.
- Use `encoding/base64` (`base64.RawURLEncoding`) directly in
  the tampering test to decode/re-encode the token.
