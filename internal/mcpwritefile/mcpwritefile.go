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

var ErrNoOutput = errors.New("no output")

var ErrPathNotInOutput = errors.New("path not in output")

func MCPWriteFile(logical_name string, path string, content string) (string, error) {
	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w: %w", ErrUnreadableFrontmatter, err)
	}

	if fm.Output == "" {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrNoOutput)
	}

	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	if fm.Output != path {
		return "", fmt.Errorf("MCPWriteFile: %w", ErrPathNotInOutput)
	}

	cfsPath := &pathutils.PathCfs{Value: path}
	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("MCPWriteFile: %w", err)
	}

	return fmt.Sprintf("wrote %s", path), nil
}
