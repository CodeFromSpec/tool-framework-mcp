// code-from-spec: ROOT/golang/server/code@PENDING
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/hash_fragment"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/load_chain"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/validate_specs"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/write_file"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const usageMessage = `Usage: framework-mcp

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
`

func main() {
	// Handle arguments.
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Print(usageMessage)
			os.Exit(0)
		}
		fmt.Fprint(os.Stderr, usageMessage)
		os.Exit(1)
	}

	// Create MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name: "framework-mcp",
	}, nil)

	// Register tools.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_chain",
		Description: "Load the spec chain context for a given logical name. Returns all relevant spec files concatenated in a single response.",
		Meta:        mcp.Meta{"anthropic/maxResultSizeChars": 500000},
	}, load_chain.HandleLoadChain)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_file",
		Description: "Write a generated source file to disk. The path must be one of the files declared in the node's outputs list. Overwrites existing content.",
	}, write_file.HandleWriteFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate the spec tree for format errors, circular references, and artifact staleness.",
	}, validate_specs.HandleValidateSpecs)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "hash_fragment",
		Description: "Calculate the hash of a line range in a file, for use in external: fragment declarations.",
	}, hash_fragment.HandleHashFragment)

	// Run server with stdio transport.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
