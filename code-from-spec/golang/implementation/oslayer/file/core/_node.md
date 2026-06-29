---
output: internal/oslayer/file.go
---

# SPEC/golang/implementation/oslayer/file/core

Handle-based file operations with automatic locking.

# Agent

Implement the type, functions, and methods listed in
the Ownership section as a Go file in package `oslayer`.

## Ownership

This file declares and implements:
- Type: `File` struct (with unexported fields)
- Functions: `OpenFile`, `RenameFile`, `DeleteFile`
- Methods: `ReadLine`, `Write`, `SkipLines`, `Close`

The following exist in other files of this package and
can be used but must not be redeclared:
- Types: `CfsPath`, `OsPath` — declared in `path.go`.
- Functions: `CfsPathToOs` — declared in `path.go`.
- Unexported functions: `fileLockShared`,
  `fileLockExclusive` — declared in platform-specific
  lock files. Call them but do not implement them.
- Error sentinels (`ErrFileUnreadable`,
  `ErrCannotCreateDirectory`, `ErrCannotOpenFile`,
  `ErrInvalidMode`, `ErrLockTimeout`, `ErrEndOfFile`,
  `ErrWrongMode`, `ErrCannotWriteFile`,
  `ErrCannotRename`, `ErrCannotDelete`) — declared in
  `errors.go`.

All unexported helpers must use the suffix `File`
(e.g. `createDirsFile`, `openStreamFile`). This is
mandatory to avoid name collisions with other files
in the package.

## Logic

### OpenFile

1. If mode is not "read", "overwrite", or "append",
   raise ErrInvalidMode.

2. Call CfsPathToOs(cfsPath). If it raises an error,
   propagate it. Store the result as os_path.

3. If mode is "read":
     Open the file for sequential reading.
     If the file cannot be opened, raise ErrFileUnreadable.
     Acquire a shared lock on the file,
     waiting up to timeoutMs milliseconds.
     If timeoutMs is zero, attempt non-blocking.
     If the lock cannot be acquired, raise ErrLockTimeout.

4. If mode is "overwrite":
     Create all intermediate directories. If any cannot
     be created, raise ErrCannotCreateDirectory.
     Open the file for writing, truncating existing
     content. If it cannot be opened, raise
     ErrCannotOpenFile.
     Acquire an exclusive lock, waiting up to timeoutMs.
     If timeoutMs is zero, attempt non-blocking.
     If the lock cannot be acquired, raise ErrLockTimeout.

5. If mode is "append":
     Create all intermediate directories. If any cannot
     be created, raise ErrCannotCreateDirectory.
     Open the file for writing without truncating. If it
     cannot be opened, raise ErrCannotOpenFile.
     Acquire an exclusive lock, waiting up to timeoutMs.
     If timeoutMs is zero, attempt non-blocking.
     If the lock cannot be acquired, raise ErrLockTimeout.

6. Return a File with mode, os_path, stream,
   closed = false, and a buffered line reader (only
   meaningful for read mode).

### ReadLine

1. If mode is not "read", raise ErrWrongMode.
2. If closed is true, raise ErrEndOfFile.
3. Read the next line up to and including the next
   newline, or until end of stream. If no more bytes,
   raise ErrEndOfFile.
4. Strip trailing line terminator: if ends with "\r\n",
   remove both; else if ends with "\n", remove it.
5. Return the resulting string.

### Write

1. If mode is not "overwrite" and not "append",
   raise ErrWrongMode.
2. Write content to the stream as UTF-8, exactly as
   received with no transformation. If the write fails,
   raise ErrCannotWriteFile.

### SkipLines

1. If mode is not "read", raise ErrWrongMode.
2. If closed is true, return immediately.
3. Repeat count times: read and discard the next line.
   If end of stream is reached before completing all
   iterations, stop without error.

### Close

1. If closed is true, return immediately.
2. Release the stream (close OS file handle and
   release lock).
3. Set closed to true.

### RenameFile

1. Call CfsPathToOs(source). If it raises an error,
   propagate it. Store as source_os.
2. Call CfsPathToOs(destination). If it raises an error,
   propagate it. Store as destination_os.
3. Perform atomic OS-level rename. If destination exists,
   overwrite it. If the rename fails, raise
   ErrCannotRename.

### DeleteFile

1. Call CfsPathToOs(cfsPath). If it raises an error,
   propagate it.
2. Delete the file at the resulting OsPath. If it cannot
   be deleted, raise ErrCannotDelete.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Normalize CRLF to LF before splitting lines.
- Use `os.OpenFile` with appropriate flags for each mode:
  read = `O_RDONLY`, overwrite = `O_WRONLY|O_CREATE|O_TRUNC`,
  append = `O_RDWR|O_CREATE|O_APPEND`.
  Append uses `O_RDWR` instead of `O_WRONLY` because on
  Windows, `O_APPEND` causes Go to replace `GENERIC_WRITE`
  with `FILE_APPEND_DATA`, which does not satisfy the
  `GENERIC_READ or GENERIC_WRITE` requirement of
  `LockFileEx`. `O_RDWR` provides `GENERIC_READ`, which
  satisfies the requirement on all platforms.
- After opening the file, call
  `fileLockShared(f, timeoutMs)` for read mode or
  `fileLockExclusive(f, timeoutMs)` for overwrite/append
  modes. These functions are defined in platform-specific
  files within the same package — do not implement them
  here.
- Use `os.Rename` for RenameFile.
- Use `os.Remove` for DeleteFile.
- Create intermediate directories with `os.MkdirAll`.
