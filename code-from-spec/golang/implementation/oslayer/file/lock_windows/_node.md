---
depends_on:
  - SPEC/golang/dependencies/golang-x-sys-windows
output: internal/oslayer/lock_windows.go
---

# SPEC/golang/implementation/oslayer/file/lock_windows

Platform-specific file locking for Windows using `LockFileEx`
with polling for timeout support.

# Agent

Generate a Go source file with build tag `//go:build windows`
in package `oslayer`.

## Ownership

This file declares and implements:
- Unexported functions: `fileLockShared`, `fileLockExclusive`

The following exist in other files of this package and can be
used but must not be redeclared:
- Error sentinels (`ErrLockTimeout`, `ErrLockFailed`) — declared
  in `errors.go`.

To avoid name collisions with other files in this package, all
identifiers you declare beyond the ones listed in the Ownership
section (functions, variables, types) must use the suffix `Win`.

## Functions to implement

```go
func fileLockShared(f *os.File, timeoutMs int) error
func fileLockExclusive(f *os.File, timeoutMs int) error
```

## Logic

Both functions follow the same pattern, differing only in the
lock flags (shared vs exclusive).

### Non-blocking path (timeoutMs <= 0)

1. Call `windows.LockFileEx` with:
   - File handle: `windows.Handle(f.Fd())`
   - Flags: `windows.LOCKFILE_FAIL_IMMEDIATELY` (and
     `windows.LOCKFILE_EXCLUSIVE_LOCK` for exclusive)
   - Reserved: 0
   - BytesLow: `^uint32(0)`, BytesHigh: `^uint32(0)` (lock
     entire file)
   - Overlapped: pointer to a zero-valued `windows.Overlapped`
2. If it succeeds, return nil.
3. If it fails, return ErrLockTimeout.

### Timeout path (timeoutMs > 0)

Same polling pattern as the Unix implementation:

1. Record the deadline: `now + timeoutMs`.
2. Set `sleep` = 1ms.
3. Loop:
   a. Call `windows.LockFileEx` with:
      - File handle: `windows.Handle(f.Fd())`
      - Flags: `windows.LOCKFILE_FAIL_IMMEDIATELY` (and
        `windows.LOCKFILE_EXCLUSIVE_LOCK` for exclusive)
      - Reserved: 0
      - BytesLow: `^uint32(0)`, BytesHigh: `^uint32(0)`
      - Overlapped: pointer to a zero-valued
        `windows.Overlapped`
   b. If it succeeds, return nil.
   c. If it returns `windows.ERROR_LOCK_VIOLATION`, the lock is
      held by another process — continue to retry.
   d. If it returns any other error, return ErrLockFailed
      (wrapping the original error).
   e. If current time >= deadline, return ErrLockTimeout.
   f. Sleep for `sleep` duration.
   g. Double `sleep`. If `sleep` > 100ms, set `sleep` = 100ms.
   h. Continue loop.

## Go-specific guidance

The file must start with:

```go
//go:build windows
```

Import `os`, `time`, `fmt`, `golang.org/x/sys/windows`.
Use `time.Sleep` for the polling interval. Use `time.Now()` and
`time.Duration` for deadline tracking.

# Private

## Decisions

### Polling instead of overlapped I/O

Changed from overlapped I/O (`CreateEvent`/`WaitForSingleObject`)
to polling with `LOCKFILE_FAIL_IMMEDIATELY` due to practical
difficulties: `LockFileEx` with overlapped I/O does not behave
like asynchronous file I/O — it blocks the calling thread instead
of returning `ERROR_IO_PENDING`, making timeout enforcement
impossible from Go.

Polling with exponential backoff is consistent with the Unix
implementation and avoids the complexity of overlapped I/O.
This decision may be worth revisiting if a reliable overlapped
locking pattern for Go on Windows is found in the future.

Backoff: 1ms, 2ms, 4ms, 8ms, 16ms, 32ms, 64ms, 100ms, 100ms,
... Cap at 100ms. Maximum overshoot past deadline is one sleep
interval (100ms).
