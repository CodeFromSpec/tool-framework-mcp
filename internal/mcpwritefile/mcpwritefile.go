// code-from-spec: ROOT/golang/implementation/mcp_tools/write_file@A67MZuzp__jCPH4wMNFqWJ-N4G8

package mcpwritefile

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var (
	// ErrUnreadableFrontmatter is returned when the node's frontmatter
	// cannot be parsed.
	ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

	// ErrNoOutputs is returned when the target node has no outputs field.
	ErrNoOutputs = errors.New("no outputs")

	// ErrPathNotInOutputs is returned when the path is not declared in
	// the node's outputs.
	ErrPathNotInOutputs = errors.New("path not in outputs")
)

// MCPWriteFile is the handler for the write_file MCP tool. It validates
// that the given path is declared in the outputs of the node identified
// by logical_name, then writes the content to that path.
//
// Parameters:
//   - logical_name: logical name of the node whose outputs authorize the write.
//   - path: relative file path from project root (forward slashes).
//   - content: complete file content (UTF-8 text).
//
// Returns a success message of the form "wrote <path>" on success.
//
// Returns an error if:
//   - the node's frontmatter cannot be parsed (ErrUnreadableFrontmatter).
//   - the node has no outputs field (ErrNoOutputs).
//   - path is not declared in the node's outputs (ErrPathNotInOutputs).
//   - the logical name cannot be resolved (LogicalNames.* errors propagated).
//   - the path fails CFS validation (PathUtils.* errors propagated).
//   - the file cannot be written (FileWriter.* errors propagated).
func MCPWriteFile(logical_name string, path string, content string) (string, error) {
	// Step 1: Resolve the logical name to a file path.
	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("resolving logical name: %w", err)
	}

	// Step 2: Parse the node's frontmatter.
	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	// Step 3: Check that the node declares outputs.
	if len(fm.Outputs) == 0 {
		return "", fmt.Errorf("%w", ErrNoOutputs)
	}

	// Step 4: Validate the path format.
	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("validating path: %w", err)
	}

	// Step 5: Check that path is declared in the node's outputs.
	found := false
	for _, output := range fm.Outputs {
		if output.Path == path {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("%w: %s", ErrPathNotInOutputs, path)
	}

	// Step 6: Write the file.
	cfsPath := &pathutils.PathCfs{Value: path}
	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	// Step 7: Return success message.
	return fmt.Sprintf("wrote %s", path), nil
}
