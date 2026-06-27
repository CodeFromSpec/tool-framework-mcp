// code-from-spec: SPEC/golang/implementation/os/file/impl@3VjL91XkWZ9bEMsQAM-CCF7NkoI

package file

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
var ErrLockTimeout = errors.New("lock timeout")
var ErrEndOfFile = errors.New("end of file")
var ErrWrongMode = errors.New("wrong mode")
var ErrCannotWriteFile = errors.New("cannot write file")
var ErrCannotRename = errors.New("cannot rename")
var ErrCannotDelete = errors.New("cannot delete")

func scanLinesNoCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		line := data[:i]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		return i + 1, line, nil
	}
	if atEOF {
		line := data
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		return len(data), line, nil
	}
	return 0, nil, nil
}

func acquireLockWithTimeout(f *os.File, shared bool, timeoutMs int) error {
	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		var err error
		if shared {
			err = fileLockShared(f)
		} else {
			err = fileLockExclusive(f)
		}
		ch <- result{err: err}
	}()

	var timeout <-chan time.Time
	if timeoutMs == 0 {
		c := make(chan time.Time)
		close(c)
		timeout = c
	} else {
		timeout = time.After(time.Duration(timeoutMs) * time.Millisecond)
	}

	select {
	case res := <-ch:
		return res.err
	case <-timeout:
		return ErrLockTimeout
	}
}

func FileOpen(cfsPath *pathutils.PathCfs, mode string, timeoutMs int) (*FileHandle, error) {
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
		if err := acquireLockWithTimeout(f, true, timeoutMs); err != nil {
			f.Close()
			if errors.Is(err, ErrLockTimeout) {
				return nil, ErrLockTimeout
			}
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		handle.stream = f
		scanner := bufio.NewScanner(f)
		scanner.Split(scanLinesNoCRLF)
		handle.scanner = scanner

	case "overwrite":
		dir := filepath.Dir(osPath.Value)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
		}
		f, err := os.OpenFile(osPath.Value, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		if err := acquireLockWithTimeout(f, false, timeoutMs); err != nil {
			f.Close()
			if errors.Is(err, ErrLockTimeout) {
				return nil, ErrLockTimeout
			}
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		handle.stream = f

	case "append":
		dir := filepath.Dir(osPath.Value)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotCreateDirectory, err)
		}
		f, err := os.OpenFile(osPath.Value, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotOpenFile, err)
		}
		if err := acquireLockWithTimeout(f, false, timeoutMs); err != nil {
			f.Close()
			if errors.Is(err, ErrLockTimeout) {
				return nil, ErrLockTimeout
			}
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
	return handle.scanner.Text(), nil
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
