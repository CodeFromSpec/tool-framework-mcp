// code-from-spec: ROOT/golang/implementation/os/file_reader@H04egMJ9MlpGY-vfRVqszoixwWk
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
	osPath  pathutils.PathOs
	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error) {
	if cfsPath == nil {
		return nil, fmt.Errorf("%w: nil cfsPath", ErrFileUnreadable)
	}

	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err.Error())
	}

	return &FileReader{
		CfsPath: *cfsPath,
		osPath:  *osPath,
		file:    f,
		scanner: bufio.NewScanner(f),
		closed:  false,
	}, nil
}

func FileReadLine(reader *FileReader) (string, error) {
	if reader == nil || reader.closed {
		return "", ErrEndOfFile
	}

	if !reader.scanner.Scan() {
		return "", ErrEndOfFile
	}

	line := reader.scanner.Text()
	line = strings.TrimRight(line, "\r")

	return line, nil
}

func FileSkipLines(reader *FileReader, count int) {
	if reader == nil || reader.closed {
		return
	}

	for i := 0; i < count; i++ {
		_, err := FileReadLine(reader)
		if errors.Is(err, ErrEndOfFile) {
			return
		}
	}
}

func FileClose(reader *FileReader) {
	if reader == nil || reader.closed {
		return
	}

	if reader.file != nil {
		_ = reader.file.Close()
	}

	reader.closed = true
}
