// code-from-spec: SPEC/golang/implementation/mcp_tools/write_file@BqvpjFnGG2snNCETdAY1a4hWLZ0
package mcpwritefile

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrQualifierNotAllowed   = errors.New("qualifier not allowed")
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput              = errors.New("no output")
var ErrPathNotInOutput       = errors.New("path not in output")

func MCPWriteFile(logicalName string, path string, content string) (string, error) {
	if logicalnames.LogicalNameHasQualifier(logicalName) {
		return "", ErrQualifierNotAllowed
	}

	nodePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		return "", fmt.Errorf("resolving logical name: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	if fm.Output == "" {
		return "", ErrNoOutput
	}

	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("validating path: %w", err)
	}

	if path != fm.Output {
		return "", ErrPathNotInOutput
	}

	cfsPath := &pathutils.PathCfs{Value: path}
	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return fmt.Sprintf("wrote %s", path), nil
}
