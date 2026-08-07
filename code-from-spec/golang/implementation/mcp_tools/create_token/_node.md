---
depends_on:
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcpcreatetoken/mcpcreatetoken.go
---

# SPEC/golang/implementation/mcp_tools/create_token

Mints an opaque token for a logical name. The
orchestrator calls this tool and hands the returned
token — never the raw logical name — to a generation
subagent, which passes it back to `load_chain` and
`write_file`. This tool is not exposed to subagents, so
a subagent can never mint a token for a node other than
the one it was dispatched for.

# Public

## Package

`package mcpcreatetoken`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpcreatetoken"`

```go
func MCPCreateToken(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. |

### Output

The opaque token string.

### Errors

- `ErrNotASpecReference`: the logical name is not a
  SPEC/ reference.
- `ErrQualifierNotAllowed`: the logical name contains a
  parenthetical qualifier.

# Agent

Implement the create token tool as a Go package.

## Logic

1. If `logicalName` does not start with `"SPEC/"`,
   return `ErrNotASpecReference`.

2. If `logicalName` contains `"("`, return
   `ErrQualifierNotAllowed`.

3. Call `subagenttoken.SubagentTokenGenerate(logicalName)`.
   If it fails, propagate the error.

4. Return the token.

## Go-specific guidance

- Use the `subagenttoken` package for
  `SubagentTokenGenerate`.
- Define sentinel errors: `ErrNotASpecReference`,
  `ErrQualifierNotAllowed`.
- The package name should be `mcpcreatetoken`.
