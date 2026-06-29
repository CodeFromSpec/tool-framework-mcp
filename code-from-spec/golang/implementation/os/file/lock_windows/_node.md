---
output: internal/file/file_lock_windows.go
---

# SPEC/golang/implementation/os/file/lock_windows

Platform-specific file locking for Windows using
overlapped I/O with real kernel-level timeout.

# Agent

Generate a Go source file with build tag `//go:build windows`
in package `file`.

Implement two unexported functions:

```go
func fileLockShared(f *os.File, timeoutMs int) error
func fileLockExclusive(f *os.File, timeoutMs int) error
```

## Logic

Both functions follow the same pattern, differing only
in the lock flags (shared vs exclusive).

### Non-blocking path (timeoutMs <= 0)

1. Call `windows.LockFileEx` with:
   - File handle: `windows.Handle(f.Fd())`
   - Flags: `windows.LOCKFILE_FAIL_IMMEDIATELY` (and
     `windows.LOCKFILE_EXCLUSIVE_LOCK` for exclusive)
   - Reserved: 0
   - BytesLow: `^uint32(0)`, BytesHigh: `^uint32(0)`
     (lock entire file)
   - Overlapped: pointer to a zero-valued
     `windows.Overlapped`

2. If it succeeds, return nil.
   If it fails, return ErrLockTimeout.

### Timeout path (timeoutMs > 0)

1. Create an event handle using
   `windows.CreateEvent(nil, 1, 0, nil)`. If it fails,
   return the error. Defer `windows.CloseHandle(event)`.

2. Build a `windows.Overlapped` struct with
   `HEvent` set to the event handle.

3. Call `windows.LockFileEx` with:
   - File handle: `windows.Handle(f.Fd())`
   - Flags: for exclusive, use
     `windows.LOCKFILE_EXCLUSIVE_LOCK`. For shared,
     use 0 (no flags). Do NOT set
     `LOCKFILE_FAIL_IMMEDIATELY`.
   - Reserved: 0
   - BytesLow: `^uint32(0)`, BytesHigh: `^uint32(0)`
   - Overlapped: pointer to the overlapped struct

4. If `LockFileEx` succeeds (returns nil), the lock
   was acquired immediately. Return nil.

5. If `LockFileEx` returns `windows.ERROR_IO_PENDING`:
   Call
   `windows.WaitForSingleObject(event, uint32(timeoutMs))`.
   If the result is `windows.WAIT_OBJECT_0`, the lock
   was acquired. Return nil.
   If the result is `windows.WAIT_TIMEOUT`, cancel the
   I/O with `windows.CancelIo(windows.Handle(f.Fd()))`
   and return ErrLockTimeout.
   For any other result, return the error.

6. For any other error from `LockFileEx`, return it.

## Go-specific guidance

The file must start with:

```go
//go:build windows
```

Import `os` and `golang.org/x/sys/windows`.
