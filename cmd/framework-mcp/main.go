// code-from-spec: ROOT/golang/implementation/server@Bjv1szpQKqSrjtiBIaYT29OtDzw
package main

import (
	"context"
	"encoding/json"
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
	type LoadChainArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"logical name of the node to load the chain for"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_chain",
		Description: "Load the spec chain context for a given logical name. Returns all relevant spec files concatenated in a single response.",
		Meta:        mcp.Meta{"anthropic/maxResultSizeChars": 500000},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args LoadChainArgs) (*mcp.CallToolResult, any, error) {
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
			content = append(content, &mcp.TextContent{Text: *result.Input})
		}

		return &mcp.CallToolResult{
			Content: content,
		}, nil, nil
	})

	// write_file tool
	type WriteFileArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"logical name of the node whose outputs list authorizes the write"`
		Path        string `json:"path"         jsonschema:"relative file path from project root"`
		Content     string `json:"content"      jsonschema:"complete file content to write"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_file",
		Description: "Write a generated source file to disk. The path must be one of the files declared in the node's outputs list. Overwrites existing content.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args WriteFileArgs) (*mcp.CallToolResult, any, error) {
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
		Description: "Validate all specs in the spec tree and check whether output files have current artifact tags.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		report := mcpvalidatespecs.MCPValidateSpecs()
		text := formatValidationReport(report)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	// hash_fragment tool
	type HashFragmentArgs struct {
		Path  string `json:"path"  jsonschema:"relative file path (forward slashes) from project root"`
		Lines string `json:"lines" jsonschema:"line range in the form start-end (e.g. 150-210), inclusive 1-based"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hash_fragment",
		Description: "Calculate the SHA-1 hash (base64url, 27 chars) of a line range within a file.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args HashFragmentArgs) (*mcp.CallToolResult, any, error) {
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
		return "Spec tree is valid and all outputs are up to date."
	}

	var sb strings.Builder

	if len(report.FormatErrors) > 0 {
		sb.WriteString("Format errors:\n")
		for _, fe := range report.FormatErrors {
			// Use JSON marshalling as a safe fallback to render the struct fields
			// without importing spectreevalidate directly.
			b, err := json.Marshal(fe)
			if err != nil {
				sb.WriteString(fmt.Sprintf("  %+v\n", fe))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", string(b)))
			}
		}
	}

	if len(report.Cycles) > 0 {
		sb.WriteString("Cycles:\n")
		for _, name := range report.Cycles {
			sb.WriteString(fmt.Sprintf("  %s\n", name))
		}
	}

	if len(report.Staleness) > 0 {
		sb.WriteString("Staleness:\n")
		for _, se := range report.Staleness {
			sb.WriteString(fmt.Sprintf("  node: %s  output: %s  path: %s  status: %s  rank: %d\n",
				se.Node, se.OutputID, se.ArtifactPath, se.Status, se.Rank))
			if se.Detail != "" {
				sb.WriteString(fmt.Sprintf("    detail: %s\n", se.Detail))
			}
		}
	}

	return sb.String()
}
