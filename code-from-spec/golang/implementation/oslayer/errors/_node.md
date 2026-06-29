---
output: internal/oslayer/errors.go
---

# SPEC/golang/implementation/oslayer/errors

All error sentinels for the oslayer package, declared
in a single file to avoid collisions across the
independently generated files in the package.

# Agent

Generate a Go file in package `oslayer` declaring all
error sentinels listed below. Use `errors.New` for
each. No logic — declarations only.

## Ownership

This file declares all error sentinels for the package.
No other file in the package may declare error
sentinels. This file has no unexported helpers.

## Declarations

```go
var (
	// Path errors
	ErrCannotDetermineRoot  = errors.New("cannot determine project root")
	ErrPathEmpty            = errors.New("path is empty")
	ErrPathAbsolute         = errors.New("path is absolute")
	ErrPathContainsBackslash = errors.New("path contains backslash")
	ErrDirectoryTraversal   = errors.New("path contains directory traversal")
	ErrResolvesOutsideRoot  = errors.New("path resolves outside project root")

	// File listing errors
	ErrDirectoryNotFound = errors.New("directory not found")
	ErrWalkError         = errors.New("error walking directory")

	// File operation errors
	ErrFileUnreadable        = errors.New("file unreadable")
	ErrCannotCreateDirectory = errors.New("cannot create directory")
	ErrCannotOpenFile        = errors.New("cannot open file")
	ErrInvalidMode           = errors.New("invalid file mode")
	ErrLockTimeout           = errors.New("lock timeout")
	ErrLockFailed            = errors.New("lock failed")
	ErrFileIO                = errors.New("file I/O error")
	ErrEndOfFile             = errors.New("end of file")
	ErrWrongMode             = errors.New("wrong file mode for this operation")
	ErrCannotWriteFile       = errors.New("cannot write file")
	ErrCannotRename          = errors.New("cannot rename file")
	ErrCannotDelete          = errors.New("cannot delete file")
)
```
