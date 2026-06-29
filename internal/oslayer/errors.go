// code-from-spec: SPEC/golang/implementation/oslayer/errors@C2aFtrhMZoU4GbLZfh8pwM_49ZU
package oslayer

import "errors"

var (
	ErrCannotDetermineRoot   = errors.New("cannot determine project root")
	ErrPathEmpty             = errors.New("path is empty")
	ErrPathAbsolute          = errors.New("path is absolute")
	ErrPathContainsBackslash = errors.New("path contains backslash")
	ErrDirectoryTraversal    = errors.New("path contains directory traversal")
	ErrResolvesOutsideRoot   = errors.New("path resolves outside project root")

	ErrDirectoryNotFound = errors.New("directory not found")
	ErrWalkError         = errors.New("error walking directory")

	ErrFileUnreadable        = errors.New("file unreadable")
	ErrCannotCreateDirectory = errors.New("cannot create directory")
	ErrCannotOpenFile        = errors.New("cannot open file")
	ErrInvalidMode           = errors.New("invalid file mode")
	ErrLockTimeout           = errors.New("lock timeout")
	ErrEndOfFile             = errors.New("end of file")
	ErrWrongMode             = errors.New("wrong file mode for this operation")
	ErrCannotWriteFile       = errors.New("cannot write file")
	ErrCannotRename          = errors.New("cannot rename file")
	ErrCannotDelete          = errors.New("cannot delete file")
)
