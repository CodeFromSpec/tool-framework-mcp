# ROOT/golang/implementation/server

Entry point: handles argument validation, creates and configures
the MCP server, registers tools, and runs the server.

# Public

## Package

`package main`

## Startup sequence

1. If `len(os.Args) > 1` and `os.Args[1]` is `--help`, `-h`, or
   `help`, print the usage message to stdout and exit 0.
2. If `len(os.Args) > 1` (any other argument), print the usage
   message to stderr and exit 1.
3. Create the MCP server via `mcp.NewServer` with
   `Implementation.Name` = `"framework-mcp"`.
4. Register tools using `mcp.AddTool`. For each tool, construct
   the `mcp.Tool` inline with the name and description from the
   corresponding tool definition spec, and pass the exported
   handler from the package:
   - `load_chain.HandleLoadChain` with `LoadChainArgs`.
     Set `Meta: mcp.Meta{"anthropic/maxResultSizeChars": 500000}`
     on the tool so that `tools/list` advertises the maximum
     result size to the client.
   - `write_file.HandleWriteFile` with `WriteFileArgs`
   - `validate_specs.HandleValidateSpecs` with `ValidateSpecsArgs`
   - `hash_fragment.HandleHashFragment` with `HashFragmentArgs`
5. Call `s.Run(context.Background(), &mcp.StdioTransport{})`.
6. If `Run` returns an error, print it to stderr and exit 1.
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
  hash_fragment    Calculate hash of a file line range.

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
