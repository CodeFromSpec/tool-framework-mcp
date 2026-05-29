// code-from-spec: ROOT/golang/implementation/os/file_writer@giSZy_mwvtZyBg1rL8Zv_QpAEak

package filewriter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var (
	// ErrCannotCreateDirectory is returned when an intermediate directory
	// cannot be created during a write operation.
	ErrCannotCreateDirectory = errors.New("cannot create directory")

	// ErrCannotWriteFile is returned when the file cannot be written.
	ErrCannotWriteFile = errors.New("cannot write file")
)

// FileWrite writes content to the file at cfs_path as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is
// created. Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line
// endings or other transformations is applied.
//
// The path is validated before writing — if validation fails, no file
// or directory is created.
//
// Possible errors:
//   - pathutils.ErrPathEmpty
//   - pathutils.ErrPathAbsolute
//   - pathutils.ErrPathContainsBackslash
//   - pathutils.ErrDirectoryTraversal
//   - pathutils.ErrResolvesOutsideRoot
//   - pathutils.ErrCannotDetermineRoot
//   - ErrCannotCreateDirectory
//   - ErrCannotWriteFile
func FileWrite(cfs_path *pathutils.PathCfs, content string) error {
	osPath, err := pathutils.PathCfsToOs(cfs_path)
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
