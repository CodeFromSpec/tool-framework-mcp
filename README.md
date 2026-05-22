# tool-framework-mcp

MCP server for [Code from Spec](https://github.com/CodeFromSpec/framework).
Provides tools for spec validation, code generation, and
artifact management.

## Tools

- **load_chain** — returns the complete spec chain for a given
  logical name, with frontmatter stripped from ancestors and
  dependencies, duplicate files removed, existing source files
  included, and the chain hash for the artifact tag
- **write_file** — writes a generated file to disk, validated
  against the node's `outputs` list
- **check** — validates the spec tree for format errors, circular
  references, and artifact staleness
- **hash_fragment** — calculates the hash of a line range in a
  file, for use in `external:` fragment declarations

## Install

Download the latest release for your platform from
[Releases](https://github.com/CodeFromSpec/tool-framework-mcp/releases)
and extract the binary into your project's `tools/` directory.

Or build from source:

```bash
go build -o tools/framework-mcp ./cmd/framework-mcp
```

## Configure

Register the server in `.claude/settings.json`:

```json
{
  "mcpServers": {
    "framework-mcp": {
      "type": "stdio",
      "command": "tools/framework-mcp"
    }
  }
}
```

On Windows, use `tools/framework-mcp.exe` as the command.

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
  check            Validate specs and check artifact staleness.
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

## Documentation

- [Code from Spec framework](https://github.com/CodeFromSpec/framework)
- [Getting Started](https://github.com/CodeFromSpec/framework/blob/main/docs/GETTING_STARTED.md)
- [Code Generation with Subagents](https://github.com/CodeFromSpec/framework/blob/main/rules/CODE_GENERATION.md)
