---
depends_on:
  - SPEC/golang/implementation/os/path_utils
output: internal/file/file.go
---

# SPEC/golang/implementation/os/file/impl

Handle-based file operations with automatic locking.

# Public

## Package

`package file`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/file"`

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

Implement the `file` package, including its interface.

## Logic

### FileOpen

1. If mode is not "read", "overwrite", or "append",
   raise ErrInvalidMode.

2. Call PathCfsToOs(cfs_path). If it raises a PathUtils
   error, propagate it. Store the result as os_path.

3. If mode is "read":
     Acquire a shared lock on the file at os_path,
     waiting up to timeout_ms milliseconds.
     If timeout_ms is zero, attempt non-blocking.
     If the lock cannot be acquired, raise ErrLockTimeout.
     Open the file for sequential reading.
     If the file cannot be opened, raise ErrFileUnreadable.

4. If mode is "overwrite":
     Create all intermediate directories. If any cannot
     be created, raise ErrCannotCreateDirectory.
     Acquire an exclusive lock, waiting up to timeout_ms.
     If timeout_ms is zero, attempt non-blocking.
     If the lock cannot be acquired, raise ErrLockTimeout.
     Open the file for writing, truncating existing
     content. If it cannot be opened, raise
     ErrCannotOpenFile.

5. If mode is "append":
     Create all intermediate directories. If any cannot
     be created, raise ErrCannotCreateDirectory.
     Acquire an exclusive lock, waiting up to timeout_ms.
     If timeout_ms is zero, attempt non-blocking.
     If the lock cannot be acquired, raise ErrLockTimeout.
     Open the file for writing without truncating. If it
     cannot be opened, raise ErrCannotOpenFile.

6. Return a FileHandle with mode, os_path, stream,
   closed = false, and a buffered line reader (only
   meaningful for read mode).

### FileReadLine

1. If handle.mode is not "read", raise ErrWrongMode.
2. If handle.closed is true, raise ErrEndOfFile.
3. Read the next line up to and including the next
   newline, or until end of stream. If no more bytes,
   raise ErrEndOfFile.
4. Strip trailing line terminator: if ends with "\r\n",
   remove both; else if ends with "\n", remove it.
5. Return the resulting string.

### FileWrite

1. If handle.mode is not "overwrite" and not "append",
   raise ErrWrongMode.
2. Write content to handle.stream as UTF-8, exactly as
   received with no transformation. If the write fails,
   raise ErrCannotWriteFile.

### FileSkipLines

1. If handle.mode is not "read", raise ErrWrongMode.
2. If handle.closed is true, return immediately.
3. Repeat count times: read and discard the next line.
   If end of stream is reached before completing all
   iterations, stop without error.

### FileClose

1. If handle.closed is true, return immediately.
2. Release handle.stream (close OS file handle and
   release lock).
3. Set handle.closed to true.

### FileRename

1. Call PathCfsToOs(source). If it raises an error,
   propagate it. Store as source_os.
2. Call PathCfsToOs(destination). If it raises an error,
   propagate it. Store as destination_os.
3. Perform atomic OS-level rename. If destination exists,
   overwrite it. If the rename fails, raise
   ErrCannotRename.

### FileDelete

1. Call PathCfsToOs(cfs_path). If it raises an error,
   propagate it.
2. Delete the file at the resulting PathOs. If it cannot
   be deleted, raise ErrCannotDelete.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.OpenFile` with appropriate flags for each mode:
  read = `O_RDONLY`, overwrite = `O_WRONLY|O_CREATE|O_TRUNC`,
  append = `O_RDWR|O_CREATE|O_APPEND`.
  Append uses `O_RDWR` instead of `O_WRONLY` because on
  Windows, `O_APPEND` causes Go to replace `GENERIC_WRITE`
  with `FILE_APPEND_DATA`, which does not satisfy the
  `GENERIC_READ or GENERIC_WRITE` requirement of
  `LockFileEx`. `O_RDWR` provides `GENERIC_READ`, which
  satisfies the requirement on all platforms.
- Normalize CRLF to LF before splitting lines.
- After opening the file, call
  `fileLockShared(f, timeoutMs)` for read mode or
  `fileLockExclusive(f, timeoutMs)` for overwrite/append
  modes. These functions are defined in platform-specific
  files within the same package — do not implement them
  here.
- Use `os.Rename` for `FileRename`.
- Use `os.Remove` for `FileDelete`.
- Create intermediate directories with `os.MkdirAll`.
