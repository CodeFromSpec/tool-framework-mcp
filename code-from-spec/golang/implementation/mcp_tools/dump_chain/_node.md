---
depends_on:
  - SPEC/golang/implementation/mcp_tools/load_chain
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcpdumpchain/mcpdumpchain.go
---

# SPEC/golang/implementation/mcp_tools/dump_chain

Saves the spec chain for a given node to a file under
`code-from-spec/.dump/` for inspection. Produces the
same document the generation subagent would receive,
allowing the orchestrator or the human to inspect it.
Each node dumps to its own file, so multiple dumps do
not overwrite each other.

# Public

## Package

`package mcpdumpchain`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpdumpchain"`

## Interface

```go
func MCPDumpChain(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. The node must declare `type`. |

### Output

A success message: `"wrote <path>"`, where `<path>`
is the dump file path derived from the logical name
(e.g. `code-from-spec/.dump/SPEC_golang_implementation_chain_hash.xml`).

### Errors

- Propagated errors from `subagenttoken`,
  `MCPLoadChain` (including `ErrNoOutput`,
  `ErrInvalidOutputPath`, `ErrArtifactModified`).
- Propagated errors from `oslayer` package.

# Agent

Implement the dump chain tool as a Go package.

## Logic

1. Call `subagenttoken.SubagentTokenGenerate(logical_name)`
   to obtain a token. If it fails, propagate the error.

2. Call `mcploadchain.MCPLoadChain(token)`. If it fails,
   propagate the error. Store the result as
   `chain_content`.

3. Derive the dump file path from `logical_name`:
   replace every "/" with "_" to form the file name,
   then join with the dump directory:
   `"code-from-spec/.dump/" + <converted> + ".xml"`.
   For example, "SPEC/golang/implementation/chain/hash"
   becomes
   "code-from-spec/.dump/SPEC_golang_implementation_chain_hash.xml".
   Store as `dump_path`.

4. Call `oslayer.OpenFile(oslayer.CfsPath(dump_path),
   "overwrite", 30000)`. If it fails, propagate the
   error. Store as handle. ("overwrite" mode creates
   the `.dump` directory if it does not exist.)

5. Call `handle.Write(chain_content)`. If it fails,
   call `handle.Close()`, then propagate the error.

6. Call `handle.Close()`.

7. Return "wrote <dump_path>".

## Go-specific guidance

- Use the `subagenttoken` package for
  `SubagentTokenGenerate`.
- Use the `mcploadchain` package for `MCPLoadChain`.
- Use the `oslayer` package for `OpenFile`, `.Write()`,
  `.Close()`, and `CfsPath`.
- Use `strings.ReplaceAll(logical_name, "/", "_")` to
  convert the logical name into the file name.
- The package name should be `mcpdumpchain`.
- The output file is under `code-from-spec/.dump/`,
  named after the target node's logical name with
  slashes replaced by underscores.
