# tool-framework-mcp

MCP server for [Code from Spec](https://github.com/CodeFromSpec/framework).
Provides tools for spec validation, code generation, and
artifact management.

## Tools

- **load_chain** — returns the complete spec chain for a given
  logical name, with artifact tag lines removed from artifact
  dependencies and input, existing source files included, and
  the chain hash for the artifact tag
- **write_file** — writes a generated file to disk, validated
  against the node's declared `output`
- **validate_specs** — validates the spec tree for format errors,
  circular references, and artifact staleness
- **chain_hash** — computes the chain hash for a node without
  assembling the full context
- **version** — returns the tool version

## Install

Download the latest release for your platform from
[Releases](https://github.com/CodeFromSpec/tool-framework-mcp/releases)
and extract the binary into your project's
`code-from-spec/_tools/` directory.

Or build from source:

```bash
go build -o code-from-spec/_tools/framework-mcp ./cmd/framework-mcp
```

## Configure

Register the server in `.mcp.json` at the project root:

```json
{
  "mcpServers": {
    "framework-mcp": {
      "type": "stdio",
      "command": "code-from-spec/_tools/framework-mcp"
    }
  }
}
```

On Windows, use `code-from-spec/_tools/framework-mcp.exe`
as the command.

## Usage

The server takes no arguments. Run `framework-mcp --help` for
usage information.

```
Usage: framework-mcp

Starts an MCP server over stdin/stdout for Code from Spec
subagents.

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

## Documentation

- [Code from Spec framework](https://github.com/CodeFromSpec/framework)

