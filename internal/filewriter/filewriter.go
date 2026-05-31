// code-from-spec: ROOT/golang/implementation/os/file_writer@k5qjSPk4qgOH2lu7F6wLjeKvLv4

package filewriter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrCannotCreateDirectory is returned when an intermediate directory cannot be created.
var ErrCannotCreateDirectory = errors.New("cannot create directory")

// ErrCannotWriteFile is returned when the file cannot be written.
var ErrCannotWriteFile = errors.New("cannot write file")

// FileWrite writes content to the file at cfsPath as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is created.
// Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line endings
// or other transformations.
//
// The path is validated before writing — if validation fails, no file or
// directory is created.
//
// Errors:
//   - ErrCannotCreateDirectory: an intermediate directory cannot be created.
//   - ErrCannotWriteFile: the file cannot be written.
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileWrite(cfsPath *pathutils.PathCfs, content string) error {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(osPath.Value)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
	}

	if err := os.WriteFile(osPath.Value, []byte(content), 0644); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotWriteFile, err)
	}

	return nil
}
