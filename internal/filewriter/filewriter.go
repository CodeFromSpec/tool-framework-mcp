// code-from-spec: ROOT/golang/implementation/os/file_writer@1fVP2wN5qSJyggs_HC2J4CMUS_g

package filewriter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrCannotCreateDirectory is returned when an intermediate directory
// cannot be created while preparing to write the file.
var ErrCannotCreateDirectory = errors.New("cannot create directory")

// ErrCannotWriteFile is returned when the file cannot be written after
// the directory structure has been prepared.
var ErrCannotWriteFile = errors.New("cannot write file")

// FileWrite writes content to the file at cfsPath as UTF-8 encoded text.
// If the file already exists it is overwritten. If it does not exist it
// is created. Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line
// endings or other transformations is applied.
//
// The path is validated before any file or directory is created. If
// validation fails, no changes are made to the filesystem.
//
// Errors:
//   - ErrCannotCreateDirectory: an intermediate directory cannot be created.
//   - ErrCannotWriteFile: the file cannot be written.
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileWrite(cfsPath *pathutils.PathCfs, content string) error {
	// Step 1: Convert cfs_path to an OS path.
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return err
	}

	// Step 2: Determine the parent directory and create it if needed.
	dir := filepath.Dir(osPath.Value)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
	}

	// Step 3: Write content to the file encoded as UTF-8 text.
	if err := os.WriteFile(osPath.Value, []byte(content), 0644); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotWriteFile, err)
	}

	return nil
}
