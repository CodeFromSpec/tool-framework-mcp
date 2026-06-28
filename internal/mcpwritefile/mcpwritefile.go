// code-from-spec: SPEC/golang/implementation/mcp_tools/write_file@5LdN_Bu9iFk6aUhBydnxT4zuNKs
package mcpwritefile

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrNotASpecReference = errors.New("not a SPEC reference")
var ErrQualifierNotAllowed = errors.New("qualifier not allowed")
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput = errors.New("no output")
var ErrPathNotInOutput = errors.New("path not in output")

func MCPWriteFile(logicalName string, path string, content string) (string, error) {
	if logicalName != "SPEC" && !strings.HasPrefix(logicalName, "SPEC/") {
		return "", ErrNotASpecReference
	}

	ln, err := logicalnames.LogicalNameParse(logicalName)
	if err != nil {
		return "", fmt.Errorf("parsing logical name: %w", err)
	}

	if ln.Qualifier != nil {
		return "", ErrQualifierNotAllowed
	}

	fm, err := frontmatter.FrontmatterParse(pathutils.PathCfs{Value: ln.Path})
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

	cfsPath := pathutils.PathCfs{Value: path}
	handle, err := file.FileOpen(cfsPath, "overwrite", 30000)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}

	if err := file.FileWrite(handle, content); err != nil {
		file.FileClose(handle)
		return "", fmt.Errorf("writing file: %w", err)
	}

	file.FileClose(handle)

	return fmt.Sprintf("wrote %s", path), nil
}
