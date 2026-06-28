---
depends_on:
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
output: internal/mcpdumpchain/mcpdumpchain.go
---

# SPEC/golang/implementation/mcp_tools/dump_chain

Saves the spec chain for a given node to a file for
inspection. Produces the same document the generation
subagent would receive, allowing the orchestrator or
the human to inspect it.

# Public

## Package

`package mcpdumpchain`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpdumpchain"`

## Interface

```go
func MCPDumpChain(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. The node must declare `output`. |

### Output

A success message: `"wrote dump_chain.xml"`.

### Errors

- Propagated errors from `MCPLoadChain` (including
  `ErrNoOutput`, `ErrInvalidOutputPath`,
  `ErrArtifactModified`).
- Propagated errors from `file` package.

# Agent

Implement the dump chain tool as a Go package.

## Logic

1. Call `MCPLoadChain(logical_name)`. If it fails,
   propagate the error. Store the result as
   `chain_content`.

2. Call `FileOpen(PathCfs{Value: "dump_chain.xml"},
   "overwrite", 30000)`. If it fails, propagate the
   error. Store as handle.

3. Call `FileWrite(handle, chain_content)`. If it
   fails, call `FileClose(handle)`, then propagate
   the error.

4. Call `FileClose(handle)`.

5. Return "wrote dump_chain.xml".

## Go-specific guidance

- Use the `mcploadchain` package for `MCPLoadChain`.
- Use the `file` package for `FileOpen`, `FileWrite`,
  `FileClose`.
- Use the `pathutils` package for `PathCfs`.
- The package name should be `mcpdumpchain`.
- The output file is always `dump_chain.xml` at the
  project root, regardless of the target node.
