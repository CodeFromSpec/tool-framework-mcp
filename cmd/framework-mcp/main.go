// code-from-spec: SPEC/golang/implementation/server@Mrw_QYq1zMCblBCt9v09_o1hePU
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpchainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpwritefile"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var Version = "dev"

const usageMessage = `Usage: framework-mcp

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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate the spec tree and check whether output artifacts are up to date.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		report := mcpvalidatespecs.MCPValidateSpecs()
		text := formatValidationReport(report)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	type chainHashArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"logical name of the node to compute the chain hash for"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "chain_hash",
		Description: "Compute the chain hash for a node.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args chainHashArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcpchainhash.MCPChainHash(args.LogicalName)
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "version",
		Description: "Print the tool version.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: Version}},
		}, nil, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

func formatValidationReport(report *mcpvalidatespecs.ValidationReport) string {
	if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
		return "Spec tree is valid and all artifacts are up to date."
	}

	var sb strings.Builder

	for _, e := range report.FormatErrors {
		fmt.Fprintf(&sb, "Format error — Node: %s | Rule: %s | Detail: %s\n", e.Node, e.Rule, e.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Fprintf(&sb, "Cycle detected involving: %s\n", name)
	}

	for _, s := range report.Staleness {
		fmt.Fprintf(&sb, "Staleness — Node: %s | Path: %s | Status: %s | Rank: %d | Detail: %s\n",
			s.Node, s.ArtifactPath, s.Status, s.Rank, s.Detail)
	}

	return sb.String()
}
