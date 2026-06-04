// code-from-spec: ROOT/golang/implementation/os/file_writer@XLhW0xpWsSNavwHPS_x56X84gvk
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

func FileWrite(cfsPath *pathutils.PathCfs, content string) error {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	dir := filepath.Dir(osPath.Value)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
	}

	err = os.WriteFile(osPath.Value, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotWriteFile, err)
	}

	return nil
}
