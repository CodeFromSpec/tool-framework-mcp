---
depends_on:
  - SPEC/functional/logic/os/path_utils
output: code-from-spec/functional/logic/os/file/output.md
---

# SPEC/functional/logic/os/file

Handle-based file operations with automatic locking.

# Public

## Namespace

    namespace: file

## Interface

```
record FileHandle
  mode: string

function FileOpen(cfs_path: pathutils.PathCfs, mode: string, timeout_ms: integer) -> FileHandle
  errors:
    - FileUnreadable: mode is "read" and the file cannot be
      opened (does not exist, permission denied, or other
      OS error).
    - CannotCreateDirectory: mode is "overwrite" or "append"
      and an intermediate directory cannot be created.
    - CannotOpenFile: mode is "overwrite" or "append" and
      the file cannot be opened for writing.
    - InvalidMode: mode is not "read", "overwrite", or
      "append".
    - LockTimeout: the lock could not be acquired within
      the specified timeout.
    - (PathUtils.*): propagated from PathCfsToOs.

function FileReadLine(handle: FileHandle) -> string
  errors:
    - EndOfFile: no more lines to read.
    - WrongMode: handle was not opened in "read" mode.

function FileWrite(handle: FileHandle, content: string)
  errors:
    - WrongMode: handle was not opened in "overwrite" or
      "append" mode.
    - CannotWriteFile: the content cannot be written.

function FileSkipLines(handle: FileHandle, count: integer)
  errors:
    - WrongMode: handle was not opened in "read" mode.

function FileClose(handle: FileHandle)

function FileRename(source: pathutils.PathCfs, destination: pathutils.PathCfs)
  errors:
    - CannotRename: the rename operation failed.
    - (PathUtils.*): propagated from PathCfsToOs.

function FileDelete(cfs_path: pathutils.PathCfs)
  errors:
    - CannotDelete: the file cannot be deleted (does not
      exist, permission denied, or other OS error).
    - (PathUtils.*): propagated from PathCfsToOs.
```

### FileOpen

Opens a file and acquires a lock based on the mode:

- `"read"` — shared lock. Opens an existing file for
  sequential line-by-line reading. The file must exist.
- `"overwrite"` — exclusive lock. Creates the file if it
  does not exist, or truncates it if it does. Creates
  intermediate directories as needed.
- `"append"` — exclusive lock. Creates the file if it does
  not exist, or opens it without truncating. Creates
  intermediate directories as needed.

The `timeout_ms` parameter controls how long to wait for
the lock. Must be zero or positive.
- Positive value: wait up to that many milliseconds.
  Raises `LockTimeout` if the lock is not acquired in time.
- Zero: non-blocking. Try to acquire the lock immediately,
  raise `LockTimeout` if it is not available.

The caller must call `FileClose` when done — failing to do
so leaks the file handle and the lock.

### FileReadLine

Reads the next line from the file, normalizes CRLF to LF,
and returns the line without the terminator. Raises
"end of file" when there are no more lines. Only valid
in "read" mode.

### FileWrite

Writes content to the file as UTF-8 encoded text. Content
is written exactly as received — no normalization of line
endings or other transformations. Only valid in "overwrite"
or "append" mode.

### FileSkipLines

Reads and discards `count` lines without returning their
content. Only valid in "read" mode.

### FileClose

Releases the lock and closes the file handle. After
`FileClose`, `FileReadLine` raises "end of file",
`FileSkipLines` does nothing, and `FileWrite` raises
"wrong mode".

### FileRename

Renames (moves) a file from `source` to `destination`.
Both paths are validated. If the destination exists, it
is overwritten.

### FileDelete

Deletes the file at `cfs_path`. The path is validated
before deletion.

# Agent

Generate pseudocode for each function in the interface.

## Implementation guidance

- Convert `cfs_path` to an OS path internally using
  the path conversion from `path_utils`.
- For "read" mode: open the file with a shared lock.
  Read from the file stream sequentially — do not load
  the entire file into memory. The reader is forward-only.
  No seeking or rewinding. Memory usage must not depend
  on file size.
- For "overwrite" mode: open the file with an exclusive
  lock, truncating any existing content. Create
  intermediate directories if they do not exist.
- For "append" mode: open the file with an exclusive
  lock, without truncating. Create intermediate
  directories if they do not exist.
- A final line without a trailing newline is still a valid
  line. The "end of file" error is raised on the next call
  after the last line.
- `FileSkipLines` past the end of the file is not an
  error — subsequent `FileReadLine` raises "end of file".
- `FileClose` must always release the lock, even if the
  file was never read or written.
- `FileRename` is an atomic operation at the OS level.
- `FileDelete` removes the file. If the file does not
  exist, raises CannotDelete.

# Private

## Dependencies

`PathCfs` comes via `depends_on` from `path_utils`.

## Decisions

### Unified handle model

Previously, file reading and file writing were separate
components (file_reader and file_writer). They are now
unified under a single handle-based model where the mode
determines which operations are available. This enables
automatic locking: read mode acquires a shared lock,
write modes acquire an exclusive lock. Every file
operation is concurrency-safe by default.

### Mode determines lock type

The mode parameter on FileOpen serves double duty: it
selects the available operations (read vs write) and
the lock type (shared vs exclusive). This is the natural
mapping — readers can share, writers need exclusivity.
