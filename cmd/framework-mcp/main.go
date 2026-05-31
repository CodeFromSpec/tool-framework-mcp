// code-from-spec: ROOT/golang/implementation/server@vSUvAZlx2WnHmpQa0jk51HKKm7Y
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcphashfragment"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
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
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Print(usageMessage)
			os.Exit(0)
		}
		fmt.Fprint(os.Stderr, usageMessage)
		os.Exit(1)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name: "framework-mcp",
	}, nil)

	// load_chain tool
	type loadChainArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"logical name of the node to load the chain for"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_chain",
		Description: "Load the spec chain context for a given logical name. Returns all relevant spec files concatenated in a single response.",
		Meta:        mcp.Meta{"anthropic/maxResultSizeChars": 500000},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args loadChainArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcploadchain.MCPLoadChain(args.LogicalName)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}

		content := []mcp.Content{
			&mcp.TextContent{Text: "chain_hash: " + result.ChainHash},
			&mcp.TextContent{Text: result.Context},
		}
		if result.Input != nil {
			content = append(content, &mcp.TextContent{Text: "--- input ---\n" + *result.Input})
		}

		return &mcp.CallToolResult{
			Content: content,
		}, nil, nil
	})

	// write_file tool
	type writeFileArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"logical name of the node whose outputs list authorizes the write"`
		Path        string `json:"path" jsonschema:"relative file path from project root"`
		Content     string `json:"content" jsonschema:"complete file content to write"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_file",
		Description: "Write a generated source file to disk. The path must be one of the files declared in the node's outputs list. Overwrites existing content.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args writeFileArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcpwritefile.MCPWriteFile(args.LogicalName, args.Path, args.Content)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	// validate_specs tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate the spec tree and check artifact staleness.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		report := mcpvalidatespecs.MCPValidateSpecs()

		var sb strings.Builder

		if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
			sb.WriteString("spec tree is valid and all outputs are up to date\n")
		} else {
			for _, e := range report.FormatErrors {
				fmt.Fprintf(&sb, "format error | node: %s | rule: %s | detail: %s\n",
					e.Node, e.Rule, e.Detail)
			}
			for _, name := range report.Cycles {
				fmt.Fprintf(&sb, "cycle detected | node: %s\n", name)
			}
			for _, s := range report.Staleness {
				fmt.Fprintf(&sb, "staleness | node: %s | output: %s | path: %s | status: %s | rank: %d | detail: %s\n",
					s.Node, s.OutputID, s.ArtifactPath, s.Status, s.Rank, s.Detail)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, nil, nil
	})

	// hash_fragment tool
	type hashFragmentArgs struct {
		Path  string `json:"path" jsonschema:"forward-slash relative path to the file"`
		Lines string `json:"lines" jsonschema:"line range in the form start-end (e.g. 150-210)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hash_fragment",
		Description: "Calculate the SHA-1 hash of a line range within a file.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args hashFragmentArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcphashfragment.MCPHashFragment(args.Path, args.Lines)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}
