// code-from-spec: SPEC/golang/implementation/server@8luDa157eFbG8PlKAliRnDHPCWE
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpaccept"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpdumpchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpwritefile"
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
  accept           Accept a modified artifact.
  dump_chain       Dump the spec chain to a file.
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

	s := mcp.NewServer(&mcp.Implementation{
		Name: "framework-mcp",
	}, nil)

	type LoadChainArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"Logical name of the target node."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "load_chain",
		Description: "Load the spec chain for a node.",
		Meta:        mcp.Meta{"anthropic/maxResultSizeChars": 500000},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args LoadChainArgs) (*mcp.CallToolResult, any, error) {
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

	type WriteFileArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"Logical name of the node whose output declares the target path."`
		Content     string `json:"content" jsonschema:"Complete file content (UTF-8 text)."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "write_file",
		Description: "Write a generated file to disk.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args WriteFileArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcpwritefile.MCPWriteFile(args.LogicalName, args.Content)
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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate specs and check artifact staleness.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		report := mcpvalidatespecs.MCPValidateSpecs()
		text := formatValidationReport(report)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	type AcceptArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"Logical name of the node whose artifact was modified."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "accept",
		Description: "Accept a modified artifact.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args AcceptArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcpaccept.MCPAccept(args.LogicalName)
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

	type DumpChainArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"Logical name of the target node."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "dump_chain",
		Description: "Dump the spec chain to a file.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DumpChainArgs) (*mcp.CallToolResult, any, error) {
		result, err := mcpdumpchain.MCPDumpChain(args.LogicalName)
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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "version",
		Description: "Print the tool version.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: Version}},
		}, nil, nil
	})

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func formatValidationReport(report mcpvalidatespecs.ValidationReport) string {
	var sb strings.Builder

	if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
		sb.WriteString("All specs are valid.\n")
		return sb.String()
	}

	if len(report.FormatErrors) > 0 {
		sb.WriteString("Format errors:\n")
		for _, fe := range report.FormatErrors {
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", fe.Node, fe.Rule, fe.Detail))
		}
	}

	if len(report.Cycles) > 0 {
		sb.WriteString("Cycles:\n")
		for _, c := range report.Cycles {
			sb.WriteString(fmt.Sprintf("  %s\n", c))
		}
	}

	if len(report.Staleness) > 0 {
		sb.WriteString("Staleness:\n")
		for _, se := range report.Staleness {
			sb.WriteString(fmt.Sprintf("  [%s] %s (%s) rank=%d", se.Status, se.Node, se.ArtifactPath, se.Rank))
			if se.Detail != "" {
				sb.WriteString(fmt.Sprintf(": %s", se.Detail))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
