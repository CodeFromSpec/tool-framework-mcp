# SPEC/golang/implementation/oslayer

Single OS abstraction package consolidating path
handling, file listing, and file operations with
locking.

# Public

## Package

`package oslayer`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"`

### Path types

```go
type CfsPath string
type OsPath string
```

`CfsPath` is a path in the Code from Spec standard
format: forward slash separator, relative to project
root, no `..` components, no drive letters, no leading
`/`, no backslashes.

`OsPath` is an absolute path in the OS's native format.
Never exposed in the framework's public API.

### Path functions

```go
func GetProjectRoot() (OsPath, error)
func ValidateStringIsCfsPath(value string) error
func CfsPathToOs(cfsPath CfsPath) (OsPath, error)
func OsPathToCfs(osPath OsPath) (CfsPath, error)
```

#### GetProjectRoot

Returns the current working directory as an `OsPath`.

#### ValidateStringIsCfsPath

Checks that a string is a valid `CfsPath`: non-empty,
relative, forward slashes only, no `..` traversal.
Returns nil if valid, an error otherwise.

#### CfsPathToOs

Converts a `CfsPath` to an absolute `OsPath` by
joining it with the project root. Validates the path
and resolves symlinks. Rejects paths that resolve
outside the project root.

#### OsPathToCfs

Converts an absolute `OsPath` to a `CfsPath` relative
to the project root. Resolves symlinks. Rejects paths
outside the project root.

### File listing

```go
func ListAllFiles(cfsPath CfsPath) ([]CfsPath, error)
```

Returns all files (not directories) found recursively
under the given directory. Results are sorted
alphabetically. If the directory exists but contains
no files, returns nil.

### File operations

```go
type File struct { /* unexported fields */ }

func OpenFile(cfsPath CfsPath, mode string, timeoutMs int) (*File, error)
func (f *File) ReadLine() (string, error)
func (f *File) Write(content string) error
func (f *File) SkipLines(count int) error
func (f *File) Close()
func RenameFile(source, destination CfsPath) error
func DeleteFile(cfsPath CfsPath) error
```

#### OpenFile

Opens a file and acquires a lock based on the mode:
- `"read"` — shared lock. File must exist.
- `"overwrite"` — exclusive lock. Creates or truncates.
  Creates intermediate directories.
- `"append"` — exclusive lock. Creates without
  truncating. Creates intermediate directories.

#### ReadLine

Reads the next line from a file opened in `"read"`
mode. Strips the trailing line terminator (LF or
CRLF). Returns `ErrEndOfFile` when no more data is
available.

#### Write

Writes content to a file opened in `"overwrite"` or
`"append"` mode. Writes bytes exactly as received,
no transformation.

#### SkipLines

Discards the next `count` lines from a file opened in
`"read"` mode. Stops silently if end of file is
reached before all lines are skipped.

#### Close

Releases the file handle and its lock. Safe to call
on an already-closed handle (no-op).

#### RenameFile

Performs an atomic OS-level rename. Overwrites the
destination if it exists.

#### DeleteFile

Deletes a file from disk.

### Errors

- `ErrCannotDetermineRoot` (GetProjectRoot)
- `ErrPathEmpty`, `ErrPathAbsolute`,
  `ErrPathContainsBackslash`, `ErrDirectoryTraversal`
  (ValidateStringIsCfsPath)
- `ErrResolvesOutsideRoot` (CfsPathToOs, OsPathToCfs)
- `ErrDirectoryNotFound`, `ErrWalkError` (ListAllFiles)
- `ErrFileUnreadable`, `ErrCannotCreateDirectory`,
  `ErrCannotOpenFile`, `ErrInvalidMode`, `ErrLockTimeout`,
  `ErrLockFailed` (OpenFile)
- `ErrEndOfFile`, `ErrWrongMode`, `ErrFileIO` (ReadLine)
- `ErrWrongMode`, `ErrCannotWriteFile` (Write)
- `ErrWrongMode`, `ErrFileIO` (SkipLines)
- `ErrCannotRename` (RenameFile)
- `ErrCannotDelete` (DeleteFile)
