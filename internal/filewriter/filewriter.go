// code-from-spec: ROOT/golang/implementation/os/file_writer@_uaNla0eifeTYoxssKFLGu8SLeo
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
		return err
	}

	parentDir := filepath.Dir(osPath.Value)

	err = os.MkdirAll(parentDir, 0755)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCannotCreateDirectory, err)
	}

	err = os.WriteFile(osPath.Value, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCannotWriteFile, err)
	}

	return nil
}
