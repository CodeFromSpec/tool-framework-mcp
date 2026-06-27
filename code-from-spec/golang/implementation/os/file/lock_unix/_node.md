---
output: internal/file/file_lock_unix.go
---

# SPEC/golang/implementation/os/file/lock_unix

Platform-specific file locking for Unix systems.

# Agent

Generate a Go source file with build tag `//go:build !windows`
in package `file`.

Implement two unexported functions:

```go
func fileLockShared(f *os.File) error
func fileLockExclusive(f *os.File) error
```

## Implementation

- `fileLockShared`: call `syscall.Flock(int(f.Fd()), syscall.LOCK_SH)`.
  Return the error from `syscall.Flock`, or nil on success.
- `fileLockExclusive`: call `syscall.Flock(int(f.Fd()), syscall.LOCK_EX)`.
  Return the error from `syscall.Flock`, or nil on success.

Both calls block until the lock is acquired.

The file must start with:

```go
//go:build !windows
```

Import only `os` and `syscall`.
