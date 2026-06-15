// code-from-spec: SPEC/golang/implementation/os/file_writer@YmkD9JQKzqNMZPrH6OO-rTuPnWA
package filewriter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotWriteFile = errors.New("cannot write file")

func FileWrite(cfsPath *pathutils.PathCfs, content string) error {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(osPath.Value)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("%w: %s", ErrCannotCreateDirectory, err)
	}

	if err := os.WriteFile(osPath.Value, []byte(content), 0644); err != nil {
		return fmt.Errorf("%w: %s", ErrCannotWriteFile, err)
	}

	return nil
}
