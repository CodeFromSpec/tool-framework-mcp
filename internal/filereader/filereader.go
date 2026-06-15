// code-from-spec: SPEC/golang/implementation/os/file_reader@ZCfCNnT7WTwYX0hApXEUXhjzACo
package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrEndOfFile = errors.New("end of file")

type FileReader struct {
	CfsPath pathutils.PathCfs
	osPath  pathutils.PathOs
	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

func FileOpen(cfsPath pathutils.PathCfs) (*FileReader, error) {
	osPath, err := pathutils.PathCfsToOs(&cfsPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(scanLinesKeepCR)

	return &FileReader{
		CfsPath: cfsPath,
		osPath:  *osPath,
		file:    f,
		scanner: scanner,
		closed:  false,
	}, nil
}

func FileReadLine(reader *FileReader) (string, error) {
	if reader == nil {
		return "", ErrEndOfFile
	}
	if reader.closed {
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
	if reader == nil {
		return
	}
	if reader.closed {
		return
	}

	for i := 0; i < count; i++ {
		if !reader.scanner.Scan() {
			return
		}
	}
}

func FileClose(reader *FileReader) {
	if reader == nil {
		return
	}
	if reader.closed {
		return
	}

	reader.file.Close()
	reader.closed = true
}

func scanLinesKeepCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			return i + 1, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}
