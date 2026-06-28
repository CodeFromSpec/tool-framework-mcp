---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/os/file/output.md
---

# SPEC/golang/interfaces/os/file

Handle-based file operations with automatic locking.

# Public

## Package

`package file`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"`

## Interface

```go
type FileHandle struct { /* unexported fields */ }

func FileOpen(cfsPath pathutils.PathCfs, mode string, timeoutMs int) (*FileHandle, error)
func FileReadLine(handle *FileHandle) (string, error)
func FileWrite(handle *FileHandle, content string) error
func FileSkipLines(handle *FileHandle, count int) error
func FileClose(handle *FileHandle)
func FileRename(source, destination pathutils.PathCfs) error
func FileDelete(cfsPath pathutils.PathCfs) error
```

### FileOpen

Opens a file and acquires a lock based on the mode:
- `"read"` — shared lock. File must exist.
- `"overwrite"` — exclusive lock. Creates or truncates.
  Creates intermediate directories.
- `"append"` — exclusive lock. Creates without truncating.
  Creates intermediate directories.

### Errors

- `ErrFileUnreadable`, `ErrCannotCreateDirectory`,
  `ErrCannotOpenFile`, `ErrInvalidMode`, `ErrLockTimeout`
  (FileOpen)
- `ErrEndOfFile`, `ErrWrongMode` (FileReadLine)
- `ErrWrongMode`, `ErrCannotWriteFile` (FileWrite)
- `ErrWrongMode` (FileSkipLines)
- `ErrCannotRename` (FileRename)
- `ErrCannotDelete` (FileDelete)
- Propagated errors from `pathutils` package.

# Agent

Generate an interface specification document listing
the package, import path, struct definition, and
function signatures.
