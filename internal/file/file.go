// code-from-spec: SPEC/golang/implementation/os/file@EbBJjjpAiwxP-pNi4MjLD0TXyjM

package file

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

type FileHandle struct {
	Mode    string
	osPath  pathutils.PathOs
	stream  *os.File
	closed  bool
	scanner *bufio.Scanner
}

var ErrFileUnreadable = errors.New("file unreadable")
var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotOpenFile = errors.New("cannot open file")
var ErrInvalidMode = errors.New("invalid mode")
var ErrEndOfFile = errors.New("end of file")
var ErrWrongMode = errors.New("wrong mode")
var ErrCannotWriteFile = errors.New("cannot write file")
var ErrCannotRename = errors.New("cannot rename")
var ErrCannotDelete = errors.New("cannot delete")

func FileOpen(cfsPath *pathutils.PathCfs, mode string) (*FileHandle, error) {
	if mode != "read" && mode != "overwrite" && mode != "append" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidMode, mode)
	}

	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	handle := &FileHandle{
		Mode:   mode,
		osPath: *osPath,
	}

	switch mode {
	case "read":
		f, err := os.OpenFile(osPath.Value, os.O_RDONLY, 0)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if err := fileLockShared(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		handle.stream = f
		handle.scanner = bufio.NewScanner(f)

	case "overwrite":
		dir := filepath.Dir(osPath.Value)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
		}
		f, err := os.OpenFile(osPath.Value, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		if err := fileLockExclusive(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		handle.stream = f

	case "append":
		dir := filepath.Dir(osPath.Value)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
		}
		f, err := os.OpenFile(osPath.Value, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		if err := fileLockExclusive(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		handle.stream = f
	}

	return handle, nil
}

func FileReadLine(handle *FileHandle) (string, error) {
	if handle == nil {
		return "", ErrWrongMode
	}
	if handle.Mode != "read" {
		return "", fmt.Errorf("%w: handle mode is %q", ErrWrongMode, handle.Mode)
	}
	if handle.closed {
		return "", ErrEndOfFile
	}
	if handle.scanner == nil {
		return "", ErrEndOfFile
	}
	if !handle.scanner.Scan() {
		return "", ErrEndOfFile
	}
	line := handle.scanner.Text()
	return line, nil
}

func FileWrite(handle *FileHandle, content string) error {
	if handle == nil {
		return ErrWrongMode
	}
	if handle.Mode != "overwrite" && handle.Mode != "append" {
		return fmt.Errorf("%w: handle mode is %q", ErrWrongMode, handle.Mode)
	}
	if handle.stream == nil {
		return fmt.Errorf("%w: file stream is nil", ErrCannotWriteFile)
	}
	_, err := handle.stream.WriteString(content)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotWriteFile, err)
	}
	return nil
}

func FileSkipLines(handle *FileHandle, count int) error {
	if handle == nil {
		return ErrWrongMode
	}
	if handle.Mode != "read" {
		return fmt.Errorf("%w: handle mode is %q", ErrWrongMode, handle.Mode)
	}
	if handle.closed {
		return nil
	}
	if handle.scanner == nil {
		return nil
	}
	for i := 0; i < count; i++ {
		if !handle.scanner.Scan() {
			break
		}
	}
	return nil
}

func FileClose(handle *FileHandle) {
	if handle == nil {
		return
	}
	if handle.closed {
		return
	}
	if handle.stream != nil {
		handle.stream.Close()
		handle.stream = nil
	}
	handle.closed = true
}

func FileRename(source *pathutils.PathCfs, destination *pathutils.PathCfs) error {
	sourceOs, err := pathutils.PathCfsToOs(source)
	if err != nil {
		return err
	}
	destinationOs, err := pathutils.PathCfsToOs(destination)
	if err != nil {
		return err
	}
	if err := os.Rename(sourceOs.Value, destinationOs.Value); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotRename, err)
	}
	return nil
}

func FileDelete(cfsPath *pathutils.PathCfs) error {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return err
	}
	if err := os.Remove(osPath.Value); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotDelete, err)
	}
	return nil
}
