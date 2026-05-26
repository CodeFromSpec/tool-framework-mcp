// code-from-spec: ROOT/golang/server/code@GPDp3q-AO-L5TpuOpt6h_haugRI

// Package main is the entry point for the framework-mcp MCP server.
// It starts an MCP server over stdin/stdout, registering all four
// tools: load_chain, write_file, validate_specs, and hash_fragment.
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

// usageMessage is printed to stdout (--help) or stderr (unexpected args).
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
	// Step 1 & 2: Handle command-line arguments.
	// --help / -h / help  → print usage to stdout and exit 0.
	// Any other argument  → print usage to stderr and exit 1.
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Print(usageMessage)
			os.Exit(0)
		}
		// Unexpected argument — startup error.
		fmt.Fprint(os.Stderr, usageMessage)
		os.Exit(1)
	}

	// Step 3: Create the MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name: "framework-mcp",
	}, nil)

	// Step 4a: Register load_chain tool.
	// The Meta entry advertises the maximum result size to the client,
	// which allows the client to allocate appropriate buffers.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_chain",
		Description: "Load the spec chain context for a given logical name. Returns all relevant spec files concatenated in a single response.",
		Meta:        mcp.Meta{"anthropic/maxResultSizeChars": 500000},
	}, load_chain.HandleLoadChain)

	// Step 4b: Register write_file tool.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_file",
		Description: "Write a generated source file to disk. The path must be one of the files declared in the node's outputs list. Overwrites existing content.",
	}, write_file.HandleWriteFile)

	// Step 4c: Register validate_specs tool.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate the spec tree for format errors, circular references, and artifact staleness.",
	}, validate_specs.HandleValidateSpecs)

	// Step 4d: Register hash_fragment tool.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hash_fragment",
		Description: "Calculate the hash of a line range in a file, for use in external: fragment declarations.",
	}, hash_fragment.HandleHashFragment)

	// Step 5: Run the server over stdio. This blocks until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		// Step 6: Server error — print to stderr and exit 1.
		fmt.Fprintf(os.Stderr, "framework-mcp: server error: %v\n", err)
		os.Exit(1)
	}

	// Step 7: Clean shutdown — exit 0 (implicit from falling off main).
}
