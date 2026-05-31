// code-from-spec: ROOT/golang/implementation/mcp_tools/write_file@7cQ6mRVv-9J8kZuRWzoDUdvbylE
package mcpwritefile

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnreadableFrontmatter is returned when the node's frontmatter cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrNoOutputs is returned when the target node has no outputs field.
var ErrNoOutputs = errors.New("no outputs")

// ErrPathNotInOutputs is returned when the requested path is not declared
// in the node's outputs.
var ErrPathNotInOutputs = errors.New("path not in outputs")

// MCPWriteFile writes content to path after verifying that the path is
// declared in the outputs of the node identified by logical_name.
//
// Steps:
//  1. Resolve logical_name to a spec file path via LogicalNameToPath.
//  2. Parse the frontmatter of that spec file.
//  3. Confirm the outputs field is present and non-empty.
//  4. Confirm path appears in the outputs list.
//  5. Validate path via PathValidateCfs.
//  6. Write content to path via FileWrite.
//
// Returns "wrote <path>" on success.
//
// Errors:
//   - ErrUnreadableFrontmatter: the node's frontmatter cannot be parsed.
//   - ErrNoOutputs: target node has no outputs field.
//   - ErrPathNotInOutputs: path is not declared in the node's outputs.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (PathUtils.*): propagated from PathValidateCfs.
//   - (FileWriter.*): propagated from FileWrite.
func MCPWriteFile(logical_name string, path string, content string) (string, error) {
	// Step 1 — Read frontmatter

	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w: %w", ErrUnreadableFrontmatter, err)
	}

	if len(fm.Outputs) == 0 {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrNoOutputs)
	}

	// Step 2 — Validate path

	cfsPath := &pathutils.PathCfs{Value: path}

	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	// Step 3 — Check path against outputs

	found := false
	for _, entry := range fm.Outputs {
		if entry.Path == path {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrPathNotInOutputs)
	}

	// Step 4 — Write file

	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	return fmt.Sprintf("wrote %s", path), nil
}
