# tool-framework-mcp

MCP server for [Code from Spec](https://github.com/CodeFromSpec/framework).
Provides tools for spec validation, chain assembly, artifact
generation, and cache management.

## Tools

- **load_chain** — assembles the complete spec chain for a
  node as an XML document, including disposition attributes
  and previous-generation content when cache is available
- **write_file** — writes a generated file to disk and
  updates the manifest
- **validate_specs** — validates the spec tree for format
  errors, circular references, and artifact staleness
- **accept** — accepts an artifact without regenerating,
  updating the manifest checksum and chain hash to match
  the current state
- **dump_chain** — writes the spec chain to `dump_chain.xml`
  for inspection
- **reconstruct_cache** — populates the cache from the
  current state of the repository
- **prune_cache** — removes unreferenced files from the cache
- **version** — returns the tool version

## Install

Download the latest release for your platform from
[Releases](https://github.com/CodeFromSpec/tool-framework-mcp/releases)
and place the binary in your project.

Or build from source:

```bash
go build -o code-from-spec/.tools/framework-mcp ./cmd/framework-mcp
```

On Windows:

```bash
go build -o code-from-spec/.tools/framework-mcp.exe ./cmd/framework-mcp
```

## Configure

Register the server in `.mcp.json` at the project root:

```json
{
  "mcpServers": {
    "framework-mcp": {
      "type": "stdio",
      "command": "code-from-spec/.tools/framework-mcp"
    }
  }
}
```

On Windows, use `code-from-spec/.tools/framework-mcp.exe` as the command.

## Usage

The server takes no arguments. Run `framework-mcp --help` for
usage information.

```
Usage: framework-mcp

Starts an MCP server over stdin/stdout for Code from Spec
projects.

Tools:
  load_chain          Load the spec chain for a node.
  write_file          Write a generated file to disk.
  validate_specs      Validate specs and check artifact staleness.
  accept              Accept a modified artifact.
  dump_chain          Dump the spec chain to a file.
  reconstruct_cache   Rebuild cache from current state.
  prune_cache         Remove unreferenced cache files.
  version             Print the tool version.

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
