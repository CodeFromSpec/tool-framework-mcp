// code-from-spec: ROOT/golang/implementation/os/file_reader@tm7Q1wgQDbYhjgKR9guosFaPs28
package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrEndOfFile = errors.New("end of file")

type FileReader struct {
	CfsPath pathutils.PathCfs
	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error) {
	if cfsPath == nil {
		return nil, fmt.Errorf("%w: nil path", ErrFileUnreadable)
	}

	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	scanner := bufio.NewScanner(f)

	return &FileReader{
		CfsPath: *cfsPath,
		file:    f,
		scanner: scanner,
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
		if err != nil {
			if errors.Is(err, ErrEndOfFile) {
				return
			}
			return
		}
	}
}

func FileClose(reader *FileReader) {
	if reader == nil || reader.closed {
		return
	}

	reader.file.Close()
	reader.closed = true
}
