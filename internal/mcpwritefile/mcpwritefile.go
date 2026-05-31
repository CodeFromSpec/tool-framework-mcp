// code-from-spec: ROOT/golang/implementation/mcp_tools/write_file@qBiHIjE4mDXLYycPieDPgc8L1e8

package mcpwritefile

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnreadableFrontmatter is returned when the node's frontmatter
// cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrNoOutputs is returned when the target node has no outputs field
// declared in its frontmatter.
var ErrNoOutputs = errors.New("node has no outputs")

// ErrPathNotInOutputs is returned when the requested path is not
// declared in the node's outputs list.
var ErrPathNotInOutputs = errors.New("path not declared in node outputs")

// MCPWriteFile writes content to the given path, provided that path is
// declared in the outputs field of the node identified by logical_name.
//
// The function resolves logical_name to the node's spec file path,
// parses its frontmatter, checks that path appears in the outputs list,
// validates the path, and finally writes the content.
//
// On success it returns the string "wrote <path>".
//
// Errors:
//   - ErrUnreadableFrontmatter: the node's frontmatter cannot be parsed.
//   - ErrNoOutputs: the target node has no outputs field.
//   - ErrPathNotInOutputs: path is not declared in the node's outputs.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (PathUtils.*): propagated from PathValidateCfs.
//   - (FileWriter.*): propagated from FileWrite.
func MCPWriteFile(logical_name string, path string, content string) (string, error) {
	// Step 1 — Read frontmatter

	// 1. Resolve logical name to node path.
	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	// 2. Parse frontmatter from the node file.
	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w: %w", ErrUnreadableFrontmatter, err)
	}

	// 3. Check that the node declares at least one output.
	if len(fm.Outputs) == 0 {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrNoOutputs)
	}

	// Step 2 — Validate path

	// 4. Construct a PathCfs record.
	cfsPath := &pathutils.PathCfs{Value: path}

	// 5. Validate the path format.
	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	// Step 3 — Check path against outputs

	// 6. Search for the path in the node's outputs list.
	found := false
	for _, output := range fm.Outputs {
		if output.Path == path {
			found = true
			break
		}
	}

	// 7. Raise error if path was not found in outputs.
	if !found {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrPathNotInOutputs)
	}

	// Step 4 — Write file

	// 8. Write the file content.
	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	// 9. Return success message.
	return fmt.Sprintf("wrote %s", path), nil
}
