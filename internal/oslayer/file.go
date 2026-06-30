package oslayer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type File struct {
	mode    string
	osPath  OsPath
	handle  *os.File
	scanner *bufio.Scanner
	closed  bool
}

func OpenFile(cfsPath CfsPath, mode string, timeoutMs int) (*File, error) {
	if mode != "read" && mode != "overwrite" && mode != "append" {
		return nil, fmt.Errorf("mode %q is not valid: %w", mode, ErrInvalidMode)
	}

	osPath, err := CfsPathToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	if mode == "read" {
		handle, err := os.OpenFile(string(osPath), os.O_RDONLY, 0)
		if err != nil {
			return nil, fmt.Errorf("cannot open file %q for reading: %w", osPath, ErrFileUnreadable)
		}
		if err := fileLockShared(handle, timeoutMs); err != nil {
			handle.Close()
			if errors.Is(err, ErrLockTimeout) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot acquire shared lock on %q: %w", osPath, ErrLockFailed)
		}
		scanner := bufio.NewScanner(handle)
		scanner.Split(scanLinesCRLFFile)
		return &File{
			mode:    mode,
			osPath:  osPath,
			handle:  handle,
			scanner: scanner,
			closed:  false,
		}, nil
	}

	dir := filepath.Dir(string(osPath))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directories for %q: %w", osPath, ErrCannotCreateDirectory)
	}

	if mode == "overwrite" {
		handle, err := os.OpenFile(string(osPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("cannot open file %q for overwrite: %w", osPath, ErrCannotOpenFile)
		}
		if err := fileLockExclusive(handle, timeoutMs); err != nil {
			handle.Close()
			if errors.Is(err, ErrLockTimeout) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot acquire exclusive lock on %q: %w", osPath, ErrLockFailed)
		}
		return &File{
			mode:   mode,
			osPath: osPath,
			handle: handle,
			closed: false,
		}, nil
	}

	handle, err := os.OpenFile(string(osPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %q for append: %w", osPath, ErrCannotOpenFile)
	}
	if err := fileLockExclusive(handle, timeoutMs); err != nil {
		handle.Close()
		if errors.Is(err, ErrLockTimeout) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot acquire exclusive lock on %q: %w", osPath, ErrLockFailed)
	}
	return &File{
		mode:   mode,
		osPath: osPath,
		handle: handle,
		closed: false,
	}, nil
}

func scanLinesCRLFFile(data []byte, atEOF bool) (advance int, token []byte, err error) {
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

func (f *File) ReadLine() (string, error) {
	if f.mode != "read" {
		return "", fmt.Errorf("file is not in read mode: %w", ErrWrongMode)
	}
	if f.closed {
		return "", ErrEndOfFile
	}
	if f.scanner.Scan() {
		return f.scanner.Text(), nil
	}
	if err := f.scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", ErrFileIO)
	}
	return "", ErrEndOfFile
}

func (f *File) Write(content string) error {
	if f.mode != "overwrite" && f.mode != "append" {
		return fmt.Errorf("file is not in write mode: %w", ErrWrongMode)
	}
	_, err := f.handle.WriteString(content)
	if err != nil {
		return fmt.Errorf("cannot write to file %q: %w", f.osPath, ErrCannotWriteFile)
	}
	return nil
}

func (f *File) SkipLines(count int) error {
	if f.mode != "read" {
		return fmt.Errorf("file is not in read mode: %w", ErrWrongMode)
	}
	if f.closed {
		return nil
	}
	for i := 0; i < count; i++ {
		if !f.scanner.Scan() {
			if err := f.scanner.Err(); err != nil {
				return fmt.Errorf("error reading file: %w", ErrFileIO)
			}
			return nil
		}
	}
	return nil
}

func (f *File) Close() {
	if f.closed {
		return
	}
	f.handle.Close()
	f.closed = true
}

func RenameFile(source, destination CfsPath) error {
	sourceOs, err := CfsPathToOs(source)
	if err != nil {
		return err
	}
	destinationOs, err := CfsPathToOs(destination)
	if err != nil {
		return err
	}
	if err := os.Rename(string(sourceOs), string(destinationOs)); err != nil {
		return fmt.Errorf("cannot rename %q to %q: %w", sourceOs, destinationOs, ErrCannotRename)
	}
	return nil
}

func DeleteFile(cfsPath CfsPath) error {
	osPath, err := CfsPathToOs(cfsPath)
	if err != nil {
		return err
	}
	if err := os.Remove(string(osPath)); err != nil {
		return fmt.Errorf("cannot delete file %q: %w", osPath, ErrCannotDelete)
	}
	return nil
}
