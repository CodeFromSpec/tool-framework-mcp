[//]: # (code-from-spec: SPEC/golang/interfaces/os/file@CUL7HCwDQTOyNqnHpx-Pk_APnPI)

# Package `file`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file`

## Struct Definitions

```go
package file

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileHandle holds the state for an open file, including its mode and
// underlying OS resources. The caller must call FileClose when done to
// release the file handle and lock.
type FileHandle struct {
	Mode string
}
```

## Error Sentinels

```go
package file

import "errors"

var ErrFileUnreadable        = errors.New("file unreadable")
var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotOpenFile        = errors.New("cannot open file")
var ErrInvalidMode           = errors.New("invalid mode")
var ErrLockTimeout           = errors.New("lock timeout")
var ErrEndOfFile             = errors.New("end of file")
var ErrWrongMode             = errors.New("wrong mode")
var ErrCannotWriteFile       = errors.New("cannot write file")
var ErrCannotRename          = errors.New("cannot rename")
var ErrCannotDelete          = errors.New("cannot delete")
```

## Function Signatures

```go
package file

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileOpen opens the file at cfsPath in the given mode and returns a FileHandle.
//
// Modes:
//   - "read"      — shared lock; opens an existing file for sequential line-by-line
//                   reading. The file must exist; returns ErrFileUnreadable otherwise.
//   - "overwrite" — exclusive lock; creates intermediate directories as needed,
//                   then creates or truncates the file.
//   - "append"    — exclusive lock; creates intermediate directories as needed,
//                   then creates or opens the file without truncating.
//
// The timeoutMs parameter controls how long to wait for the lock:
//   - Positive value: wait up to that many milliseconds; returns ErrLockTimeout
//                     if the lock is not acquired in time.
//   - Zero: non-blocking; returns ErrLockTimeout immediately if the lock is
//           not available.
//
// The caller must call FileClose when done — failing to do so leaks the file
// handle and lock.
//
// Returns ErrInvalidMode if mode is not one of the three values above.
// Returns ErrFileUnreadable if mode is "read" and the file cannot be opened.
// Returns ErrCannotCreateDirectory if intermediate directories cannot be created
// (modes "overwrite" and "append" only).
// Returns ErrCannotOpenFile if the file cannot be opened for writing
// (modes "overwrite" and "append" only).
// Returns ErrLockTimeout if the lock could not be acquired within timeoutMs.
// Errors from pathutils.PathCfsToOs are propagated as-is.
func FileOpen(cfsPath *pathutils.PathCfs, mode string, timeoutMs int) (*FileHandle, error)

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the line terminator.
// Returns ErrEndOfFile when there are no more lines, or after FileClose has
// been called.
// Returns ErrWrongMode if the handle was not opened in "read" mode.
func FileReadLine(handle *FileHandle) (string, error)

// FileWrite writes content to the file as UTF-8 encoded text. Content is
// written exactly as received with no line-ending normalization.
// Returns ErrWrongMode if the handle was not opened in "overwrite" or "append" mode.
// Returns ErrCannotWriteFile if the content cannot be written.
func FileWrite(handle *FileHandle, content string) error

// FileSkipLines reads and discards count lines without returning their content.
// Does nothing if the handle has been closed.
// Returns ErrWrongMode if the handle was not opened in "read" mode.
func FileSkipLines(handle *FileHandle, count int) error

// FileClose releases the lock and closes the file handle. After FileClose,
// FileReadLine returns ErrEndOfFile, FileSkipLines does nothing, and
// FileWrite returns ErrWrongMode.
func FileClose(handle *FileHandle)

// FileRename renames (moves) the file at source to destination. Both paths
// are validated. If the destination exists, it is overwritten.
// Returns ErrCannotRename if the rename operation fails.
// Errors from pathutils.PathCfsToOs are propagated as-is.
func FileRename(source *pathutils.PathCfs, destination *pathutils.PathCfs) error

// FileDelete deletes the file at cfsPath. The path is validated before deletion.
// Returns ErrCannotDelete if the file cannot be deleted.
// Errors from pathutils.PathCfsToOs are propagated as-is.
func FileDelete(cfsPath *pathutils.PathCfs) error
```

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	src := &pathutils.PathCfs{Value: "SPEC/myproject/draft.md"}
	dst := &pathutils.PathCfs{Value: "SPEC/myproject/final.md"}

	writeHandle, err := file.FileOpen(src, "overwrite", 500)
	if err != nil {
		log.Fatal(err)
	}
	if err := file.FileWrite(writeHandle, "# Final\n"); err != nil {
		log.Fatal(err)
	}
	file.FileClose(writeHandle)

	if err := file.FileRename(src, dst); err != nil {
		log.Fatal(err)
	}

	readHandle, err := file.FileOpen(dst, "read", 500)
	if err != nil {
		log.Fatal(err)
	}
	defer file.FileClose(readHandle)

	if err := file.FileSkipLines(readHandle, 1); err != nil {
		log.Fatal(err)
	}

	for {
		line, err := file.FileReadLine(readHandle)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(line)
	}

	stale := &pathutils.PathCfs{Value: "SPEC/myproject/old.md"}
	if err := file.FileDelete(stale); err != nil {
		log.Fatal(err)
	}
}
```
