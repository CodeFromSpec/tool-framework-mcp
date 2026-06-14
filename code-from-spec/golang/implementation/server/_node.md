---
depends_on:
  - SPEC/golang/dependencies/mcp-go-sdk
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain
  - ARTIFACT/golang/interfaces/mcp_tools/write_file
  - ARTIFACT/golang/interfaces/mcp_tools/validate_specs
  - ARTIFACT/golang/interfaces/mcp_tools/chain_hash
output: cmd/framework-mcp/main.go
---

# SPEC/golang/implementation/server

Entry point: handles argument validation, creates and
configures the MCP server, registers tools, and runs
the server.

# Public

## Package

`package main`

## Startup sequence

1. If `len(os.Args) > 1` and `os.Args[1]` is `--help`,
   `-h`, or `help`, print the usage message to stdout
   and exit 0.
2. If `len(os.Args) > 1` (any other argument), print
   the usage message to stderr and exit 1.
3. Create the MCP server via `mcp.NewServer` with
   `Implementation.Name` = `"framework-mcp"`.
4. Register tools using `mcp.AddTool`. For each tool,
   construct the `mcp.Tool` inline with the name and
   description, and call the exported function from the
   corresponding package:
   - `mcploadchain.MCPLoadChain` — tool name
     `load_chain`. Set `Meta:
     mcp.Meta{"anthropic/maxResultSizeChars": 500000}`
     so that `tools/list` advertises the maximum result
     size to the client.
   - `mcpwritefile.MCPWriteFile` — tool name
     `write_file`.
   - `mcpvalidatespecs.MCPValidateSpecs` — tool name
     `validate_specs`.
   - `mcpchainhash.MCPChainHash` — tool name
     `chain_hash`.
   - `version` — tool name `version`. Takes no parameters.
     Returns the value of a package-level variable
     `var Version = "dev"`. This variable is overridden
     at build time via `-ldflags
     "-X main.Version=<version>"`.
5. Call `s.Run(context.Background(), &mcp.StdioTransport{})`.
6. If `Run` returns an error, print it to stderr and
   exit 1.
7. Otherwise exit 0.

## Usage message

```
Usage: framework-mcp

Starts an MCP server over stdin/stdout for Code from Spec
projects.

Tools:
  load_chain       Load the spec chain for a node.
  write_file       Write a generated file to disk.
  validate_specs   Validate specs and check artifact staleness.
  chain_hash       Compute the chain hash for a node.
  version          Print the tool version.

MCP configuration example:
  {
    "mcpServers": {
      "framework-mcp": {
        "type": "stdio",
        "command": "<path-to-binary>"
      }
    }
  }
```

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Clean shutdown. |
| 1 | Startup error or server error. |

# Agent

## Go-specific guidance

- Import the four MCP tool packages:
  `mcploadchain`, `mcpwritefile`, `mcpvalidatespecs`,
  `mcpchainhash`.
- Each tool handler receives MCP request parameters and
  calls the corresponding package function.
- The handler wraps the function result into an MCP
  tool response (text content).
- For `MCPLoadChain`, `MCPWriteFile`, and
  `MCPChainHash`, the result is a string — return
  directly as text content.
- For `MCPValidateSpecs`, the result is
  `ValidationReport` — format as human-readable text.
- For `version`, return `Version` directly as text
  content. No external package needed — the handler
  is inline.
