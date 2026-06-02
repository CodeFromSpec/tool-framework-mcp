// code-from-spec: ROOT/golang/implementation/server@oNYmkB7A7BxxDPxMC5fFHW-W_f0
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
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
		content := []mcp.Content{
			&mcp.TextContent{Text: result.ChainHash},
			&mcp.TextContent{Text: result.Context},
		}
		if result.Input != nil {
			content = append(content, &mcp.TextContent{Text: *result.Input})
		}
		return &mcp.CallToolResult{Content: content}, nil, nil
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
		msg, err := mcpwritefile.MCPWriteFile(args.LogicalName, args.Path, args.Content)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate the spec tree format, detect dependency cycles, and check artifact staleness.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		report := mcpvalidatespecs.MCPValidateSpecs()

		if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "spec tree is valid and all artifacts are fresh"}},
			}, nil, nil
		}

		var sb strings.Builder
		for _, e := range report.FormatErrors {
			sb.WriteString(fmt.Sprintf("format error: node=%s rule=%s detail=%s\n", e.Node, e.Rule, e.Detail))
		}
		for _, name := range report.Cycles {
			sb.WriteString(fmt.Sprintf("cycle: %s\n", name))
		}
		for _, s := range report.Staleness {
			sb.WriteString(fmt.Sprintf("staleness: node=%s path=%s status=%s rank=%d detail=%s\n",
				s.Node, s.ArtifactPath, s.Status, s.Rank, s.Detail))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, nil, nil
	})

	type chainHashArgs struct {
		LogicalName string `json:"logical_name" jsonschema:"logical name of the node to compute the chain hash for"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "chain_hash",
		Description: "Compute the 27-character base64url chain hash for a given logical name.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args chainHashArgs) (*mcp.CallToolResult, any, error) {
		hash, err := mcpchainhash.MCPChainHash(args.LogicalName)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: hash}},
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
