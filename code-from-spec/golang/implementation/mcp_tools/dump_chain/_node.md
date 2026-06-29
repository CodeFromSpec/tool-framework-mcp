---
depends_on:
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/oslayer(interface)
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
- Propagated errors from `oslayer` package.

# Agent

Implement the dump chain tool as a Go package.

## Logic

1. Call `mcploadchain.MCPLoadChain(logical_name)`. If it fails,
   propagate the error. Store the result as
   `chain_content`.

2. Call `oslayer.OpenFile(oslayer.CfsPath("dump_chain.xml"),
   "overwrite", 30000)`. If it fails, propagate the
   error. Store as handle.

3. Call `handle.Write(chain_content)`. If it fails,
   call `handle.Close()`, then propagate the error.

4. Call `handle.Close()`.

5. Return "wrote dump_chain.xml".

## Go-specific guidance

- Use the `mcploadchain` package for `MCPLoadChain`.
- Use the `oslayer` package for `OpenFile`, `.Write()`,
  `.Close()`, and `CfsPath`.
- The package name should be `mcpdumpchain`.
- The output file is always `dump_chain.xml` at the
  project root, regardless of the target node.
