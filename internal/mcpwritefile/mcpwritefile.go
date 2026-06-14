// code-from-spec: ROOT/golang/implementation/mcp_tools/write_file@pdtQ1Jflp3qSpqA-OlAtySV7j7I
package mcpwritefile

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrQualifierNotAllowed = errors.New("logical name contains a parenthetical qualifier")
var ErrUnreadableFrontmatter = errors.New("node frontmatter cannot be parsed")
var ErrNoOutput = errors.New("target node has no output field")
var ErrPathNotInOutput = errors.New("path is not declared in the node's output")

func MCPWriteFile(logical_name string, path string, content string) (string, error) {
	if logicalnames.LogicalNameHasQualifier(logical_name) {
		return "", ErrQualifierNotAllowed
	}

	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	if fm.Output == "" {
		return "", ErrNoOutput
	}

	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if path != fm.Output {
		return "", ErrPathNotInOutput
	}

	cfsPath := &pathutils.PathCfs{Value: path}
	if err := filewriter.FileWrite(cfsPath, content); err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return "wrote " + path, nil
}
