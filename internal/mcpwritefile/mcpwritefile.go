// code-from-spec: ROOT/golang/implementation/mcp_tools/write_file@5-vZMzq0ehJxfT_rjLhCUykYcKs
package mcpwritefile

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

var ErrNoOutputs = errors.New("no outputs")

var ErrPathNotInOutputs = errors.New("path not in outputs")

func MCPWriteFile(logical_name string, path string, content string) (string, error) {
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

	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	found := false
	for _, output := range fm.Outputs {
		if output.Path == path {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrPathNotInOutputs)
	}

	cfsPath := &pathutils.PathCfs{Value: path}
	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	return fmt.Sprintf("wrote %s", path), nil
}
