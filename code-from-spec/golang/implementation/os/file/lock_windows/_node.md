---
output: internal/file/file_lock_windows.go
---

# SPEC/golang/implementation/os/file/lock_windows

Platform-specific file locking for Windows.

# Agent

Generate a Go source file with build tag `//go:build windows`
in package `file`.

Implement two unexported functions:

```go
func fileLockShared(f *os.File) error
func fileLockExclusive(f *os.File) error
```

## Implementation

Use `LockFileEx` from the Windows API via
`golang.org/x/sys/windows`:

- `fileLockShared`: call `windows.LockFileEx` with
  `windows.LOCKFILE_FAIL_IMMEDIATELY` cleared (blocking)
  and `windows.LOCKFILE_EXCLUSIVE_LOCK` cleared (shared).
  Lock the entire file (offset 0, length `^uint32(0)`).
  Return the error, or nil on success.
- `fileLockExclusive`: call `windows.LockFileEx` with
  `windows.LOCKFILE_EXCLUSIVE_LOCK` set.
  Lock the entire file (offset 0, length `^uint32(0)`).
  Return the error, or nil on success.

The file must start with:

```go
//go:build windows
```

Import `os` and `golang.org/x/sys/windows`.
