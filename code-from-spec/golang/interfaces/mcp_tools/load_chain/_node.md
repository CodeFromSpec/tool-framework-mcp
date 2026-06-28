---
depends_on:
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
output: code-from-spec/golang/interfaces/mcp_tools/load_chain/output.md
---

# SPEC/golang/interfaces/mcp_tools/load_chain

Loads the complete spec chain for a given node and
returns everything the subagent needs in a single
formatted string.

# Public

## Package

`package mcploadchain`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain"`

## Interface

```go
func MCPLoadChain(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. |

### Output

A single string with sections separated by delimiter
lines. The format is:

```
chain_hash: <27-character hash>
--- context ---
<context content>
--- input ---
<input content>
--- existing artifact ---
<existing artifact content>
```

The `--- input ---` section is only present when the
target node's frontmatter has a non-empty `input` field.

The `--- existing artifact ---` section is only present
when the output file exists on disk and is readable.

### Errors

- `ErrNoOutput`: target node has no output field.
- `ErrInvalidOutputPath`: the output path fails path
  validation.
- Propagated errors from `logicalnames`, `chainresolver`,
  `chainhash`, `parsenode`, `file` packages.

# Agent

Generate an interface specification document listing
the package, import path, and function signatures.
