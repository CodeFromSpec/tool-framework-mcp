// code-from-spec: ROOT/golang/implementation/os/file_reader@5IQrGChGNsogWRfbLFUFWWBSJOI
package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file cannot be opened")
var ErrEndOfFile = errors.New("end of file")

type FileReader struct {
	CfsPath pathutils.PathCfs
	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error) {
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	scanner := bufio.NewScanner(f)

	return &FileReader{
		CfsPath: *cfs_path,
		file:    f,
		scanner: scanner,
		closed:  false,
	}, nil
}

func FileReadLine(reader *FileReader) (string, error) {
	if reader.closed {
		return "", ErrEndOfFile
	}

	if !reader.scanner.Scan() {
		if err := reader.scanner.Err(); err != nil {
			return "", fmt.Errorf("%w: %s", ErrEndOfFile, err)
		}
		return "", ErrEndOfFile
	}

	line := reader.scanner.Text()
	line = strings.TrimRight(line, "\r")
	return line, nil
}

func FileSkipLines(reader *FileReader, count int) {
	if reader.closed {
		return
	}

	for i := 0; i < count; i++ {
		_, err := FileReadLine(reader)
		if err != nil {
			return
		}
	}
}

func FileClose(reader *FileReader) {
	if reader.closed {
		return
	}
	reader.file.Close()
	reader.closed = true
}
