// code-from-spec: ROOT/golang/implementation/server@shNTWyWxx4rQbmHkJqZGsqjZsmA
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

	// Register load_chain tool
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

		contents := []mcp.Content{
			&mcp.TextContent{Text: "chain_hash: " + result.ChainHash},
			&mcp.TextContent{Text: result.Context},
		}
		if result.Input != nil {
			contents = append(contents, &mcp.TextContent{Text: "--- input ---\n" + *result.Input})
		}

		return &mcp.CallToolResult{
			Content: contents,
		}, nil, nil
	})

	// Register write_file tool
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

	// Register validate_specs tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate all spec nodes and check artifact staleness across the entire spec tree.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		report := mcpvalidatespecs.MCPValidateSpecs()

		text := formatValidationReport(report)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	// Register hash_fragment tool
	type hashFragmentArgs struct {
		Path  string `json:"path" jsonschema:"file path relative to the project root using forward slashes"`
		Lines string `json:"lines" jsonschema:"line range in the form start-end (e.g. 150-210)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hash_fragment",
		Description: "Calculate the SHA-1 hash (base64url, no padding) of a line range within a file.",
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
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// formatValidationReport converts a ValidationReport into human-readable text.
func formatValidationReport(report *mcpvalidatespecs.ValidationReport) string {
	if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
		return "spec tree is valid and all artifacts are up to date"
	}

	result := ""

	if len(report.FormatErrors) > 0 {
		result += fmt.Sprintf("Format errors (%d):\n", len(report.FormatErrors))
		for _, fe := range report.FormatErrors {
			result += fmt.Sprintf("  node=%s rule=%s detail=%s\n", fe.Node, fe.Rule, fe.Detail)
		}
	}

	if len(report.Cycles) > 0 {
		result += fmt.Sprintf("Cycles (%d):\n", len(report.Cycles))
		for _, name := range report.Cycles {
			result += fmt.Sprintf("  node=%s\n", name)
		}
	}

	if len(report.Staleness) > 0 {
		result += fmt.Sprintf("Stale or missing artifacts (%d):\n", len(report.Staleness))
		for _, se := range report.Staleness {
			result += fmt.Sprintf("  node=%s output=%s path=%s status=%s rank=%d detail=%s\n",
				se.Node, se.OutputID, se.ArtifactPath, se.Status, se.Rank, se.Detail)
		}
	}

	// Append JSON representation for structured access
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err == nil {
		result += "\nJSON:\n" + string(jsonBytes)
	}

	return result
}
