// code-from-spec: ROOT/golang/internal/tools/write_file/code@mI3Fz8UagXT8P6btvmqnlX42Lqo

// Package write_file implements the write_file MCP tool.
//
// The tool accepts a logical name (ROOT/ reference), a relative file path,
// and file content. It validates that the path is declared in the node's
// outputs frontmatter, performs path safety checks, creates any missing
// intermediate directories, and writes the file to disk.
//
// Each call is fully stateless: inputs are resolved and validated
// independently every time the handler is invoked.
package write_file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WriteFileArgs holds the input parameters for the write_file tool.
// Field names match the JSON schema expected by the MCP host.
type WriteFileArgs struct {
	// LogicalName is the ROOT/ reference identifying the node whose outputs
	// list authorizes this write.
	LogicalName string `json:"logical_name" jsonschema:"Logical name of the node whose outputs list authorizes the write."`

	// Path is the relative file path from the project root where the content
	// will be written.
	Path string `json:"path" jsonschema:"Relative file path from project root."`

	// Content is the complete file content (UTF-8) to write.
	Content string `json:"content" jsonschema:"Complete file content to write."`
}

// HandleWriteFile is the MCP tool handler for the write_file tool.
//
// It validates the logical name, resolves the node's frontmatter, confirms
// the target path is declared in the outputs list, performs safety checks on
// the path, then writes the file to disk.
//
// All expected error conditions are returned as MCP tool errors (IsError: true).
// The Go error return is reserved for catastrophic server failures only.
func HandleWriteFile(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args WriteFileArgs,
) (*mcp.CallToolResult, any, error) {

	// ------------------------------------------------------------------
	// Step 1 — Validate the logical name.
	//
	// Only ROOT/ references are valid. ARTIFACT/ references and empty
	// strings are rejected here before any I/O is attempted.
	// ------------------------------------------------------------------
	if !strings.HasPrefix(args.LogicalName, "ROOT/") {
		return toolError(fmt.Sprintf(
			"invalid logical name: %q is not a recognized ROOT/ reference",
			args.LogicalName,
		)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 2 — Resolve the node file path from the logical name.
	//
	// PathFromLogicalName strips any parenthetical qualifier and converts
	// the ROOT/ reference to a file path like "code-from-spec/x/y/_node.md".
	// ------------------------------------------------------------------
	nodePath, ok := logicalnames.PathFromLogicalName(args.LogicalName)
	if !ok {
		return toolError(fmt.Sprintf(
			"invalid logical name: %q could not be resolved to a node file path",
			args.LogicalName,
		)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 3 — Read and parse the node's frontmatter.
	//
	// The node file must exist and have a non-empty outputs list.
	// We distinguish between read errors (node not found) and parse errors
	// (malformed frontmatter) to give the agent actionable messages.
	// ------------------------------------------------------------------
	fm, err := frontmatter.ParseFrontmatter(nodePath)
	if err != nil {
		if errors.Is(err, frontmatter.ErrRead) {
			// File does not exist or cannot be read — logical name points
			// to a node that does not exist in the repository.
			return toolError(fmt.Sprintf(
				"invalid logical name: node file not found for %q (resolved path: %s)",
				args.LogicalName, nodePath,
			)), nil, nil
		}
		// Parse error — the file exists but its frontmatter is malformed.
		return toolError(fmt.Sprintf(
			"invalid logical name: could not parse frontmatter for %q: %v",
			args.LogicalName, err,
		)), nil, nil
	}

	// The node must declare at least one output.
	if len(fm.Outputs) == 0 {
		return toolError(fmt.Sprintf(
			"no outputs: node %q has no outputs declared",
			args.LogicalName,
		)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 4 — Validate the write path for safety.
	//
	// ValidatePath checks for empty paths, absolute paths, directory
	// traversal sequences, and paths that resolve outside the project root.
	// The project root is the current working directory (the server always
	// runs from the project root per the project constraints).
	// ------------------------------------------------------------------
	projectRoot, err := os.Getwd()
	if err != nil {
		// Cannot determine project root — this is a server-level failure.
		return toolError(fmt.Sprintf(
			"internal error: could not determine project root: %v", err,
		)), nil, nil
	}

	if err := pathvalidation.ValidatePath(args.Path, projectRoot); err != nil {
		return toolError(fmt.Sprintf("path validation failure: %v", err)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 5 — Confirm the path is declared in the node's outputs.
	//
	// This is the security boundary: only paths explicitly listed in the
	// node's frontmatter are permitted. Comparison uses filepath.ToSlash
	// on both sides so that Windows backslash paths do not bypass the check.
	// ------------------------------------------------------------------
	normalizedInputPath := filepath.ToSlash(args.Path)
	declared := false
	for _, output := range fm.Outputs {
		if filepath.ToSlash(output.Path) == normalizedInputPath {
			declared = true
			break
		}
	}
	if !declared {
		return toolError(fmt.Sprintf(
			"path not in outputs: %q is not declared in the outputs of %q",
			args.Path, args.LogicalName,
		)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 6 — Create intermediate directories if needed.
	//
	// MkdirAll is a no-op if the directory already exists.
	// ------------------------------------------------------------------
	dir := filepath.Dir(args.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return toolError(fmt.Sprintf(
			"directory creation failure: could not create directories for %q: %v",
			args.Path, err,
		)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 7 — Write the file.
	//
	// os.WriteFile creates the file if it does not exist, or overwrites it
	// if it does. Content is treated as UTF-8 bytes.
	// ------------------------------------------------------------------
	if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
		return toolError(fmt.Sprintf(
			"write failure: could not write to %q: %v",
			args.Path, err,
		)), nil, nil
	}

	// ------------------------------------------------------------------
	// Step 8 — Return success.
	// ------------------------------------------------------------------
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("wrote %s", args.Path),
		}},
	}, nil, nil
}

// toolError constructs an MCP tool error result with the given message.
// The server continues running after returning a tool error — IsError marks
// the result as a tool-level error, not a server panic.
func toolError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: message}},
		IsError: true,
	}
}
