// code-from-spec: ROOT/golang/implementation/os/file_writer@Jg5Iup_ibh3DxFueCcqp6pu9zfY
package filewriter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotWriteFile = errors.New("cannot write file")

func FileWrite(cfs_path *pathutils.PathCfs, content string) error {
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(osPath.Value)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %s", ErrCannotCreateDirectory, err)
	}

	if err := os.WriteFile(osPath.Value, []byte(content), 0644); err != nil {
		return fmt.Errorf("%w: %s", ErrCannotWriteFile, err)
	}

	return nil
}
