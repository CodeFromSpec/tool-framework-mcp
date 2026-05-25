// code-from-spec: ROOT/golang/internal/tools/write_file/code@PENDING
package write_file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WriteFileArgs defines the input parameters for the write_file tool.
type WriteFileArgs struct {
	LogicalName string `json:"logical_name" jsonschema:"Logical name of the node whose outputs list authorizes the write."`
	Path        string `json:"path" jsonschema:"Relative file path from project root."`
	Content     string `json:"content" jsonschema:"Complete file content to write."`
}

// HandleWriteFile validates the path against the node's outputs list
// and writes the file to disk.
func HandleWriteFile(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args WriteFileArgs,
) (*mcp.CallToolResult, any, error) {
	// Step 1: Validate logical name starts with ROOT/.
	if !strings.HasPrefix(args.LogicalName, "ROOT/") && args.LogicalName != "ROOT" {
		return toolError("invalid logical name"), nil, nil
	}

	// Step 2: Resolve logical name to file path and parse frontmatter.
	nodePath, ok := logicalnames.PathFromLogicalName(args.LogicalName)
	if !ok {
		return toolError("invalid logical name"), nil, nil
	}

	fm, err := frontmatter.ParseFrontmatter(nodePath)
	if err != nil {
		return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
	}

	// Step 3: Check outputs not empty.
	if len(fm.Outputs) == 0 {
		return toolError("no outputs"), nil, nil
	}

	// Step 4: Normalize and validate the path.
	normalizedPath := filepath.ToSlash(args.Path)

	if err := pathvalidation.ValidatePath(normalizedPath, "."); err != nil {
		return toolError(fmt.Sprintf("path validation failure: %v", err)), nil, nil
	}

	// Step 5: Check path is in outputs.
	found := false
	for _, out := range fm.Outputs {
		if filepath.ToSlash(out.Path) == normalizedPath {
			found = true
			break
		}
	}
	if !found {
		return toolError("path not in outputs"), nil, nil
	}

	// Step 6: Create directories.
	fullPath := filepath.Join(".", normalizedPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return toolError(fmt.Sprintf("directory creation failure: %v", err)), nil, nil
	}

	// Step 7: Write file.
	if err := os.WriteFile(fullPath, []byte(args.Content), 0644); err != nil {
		return toolError(fmt.Sprintf("write failure: %v", err)), nil, nil
	}

	// Step 8: Return success.
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("wrote %s", normalizedPath)}},
	}, nil, nil
}

// toolError returns a CallToolResult with IsError set to true.
func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
