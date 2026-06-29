---
output: internal/oslayer/lock_unix.go
---

# SPEC/golang/implementation/oslayer/file/lock_unix

Platform-specific file locking for Unix systems using
`flock` with polling for timeout support.

# Agent

Generate a Go source file with build tag
`//go:build !windows` in package `oslayer`.

## Ownership

This file declares and implements:
- Unexported functions: `fileLockShared`,
  `fileLockExclusive`

The following exist in other files of this package and
can be used but must not be redeclared:
- Error sentinels (`ErrLockTimeout`) — declared in
  `errors.go`.

All unexported helpers must use the suffix `Lock`
(e.g. `retryWithBackoffLock`). This is mandatory to
avoid name collisions with other files in the package.

## Functions to implement

```go
func fileLockShared(f *os.File, timeoutMs int) error
func fileLockExclusive(f *os.File, timeoutMs int) error
```

## Logic

Both functions follow the same pattern, differing only
in the flock flag (`syscall.LOCK_SH` vs
`syscall.LOCK_EX`).

### Non-blocking path (timeoutMs <= 0)

1. Call `syscall.Flock(int(f.Fd()), flag|syscall.LOCK_NB)`.
2. If it succeeds, return nil.
3. If it returns `EWOULDBLOCK`, return ErrLockTimeout.
4. For any other error, return it.

### Timeout path (timeoutMs > 0)

1. Record the deadline: `now + timeoutMs`.
2. Set `sleep` = 1ms.
3. Loop:
   a. Call `syscall.Flock(int(f.Fd()), flag|syscall.LOCK_NB)`.
   b. If it succeeds, return nil.
   c. If it returns any error other than `EWOULDBLOCK`,
      return it.
   d. If current time >= deadline, return ErrLockTimeout.
   e. Sleep for `sleep` duration.
   f. Double `sleep`. If `sleep` > 100ms, set
      `sleep` = 100ms.
   g. Continue loop.

## Go-specific guidance

The file must start with:

```go
//go:build !windows
```

Import `os`, `syscall`, and `time`.
Use `time.Sleep` for the polling interval.
Use `time.Now()` and `time.Duration` for deadline
tracking.

# Private

## Decisions

### Polling with LOCK_NB instead of blocking flock

Unix `flock` has no native timeout. Blocking `flock`
cannot be interrupted safely (closing the fd from
another goroutine causes fd-reuse races). Polling
with `LOCK_NB` and exponential backoff is the
standard approach (used by SQLite, Git).

Backoff: 1ms, 2ms, 4ms, 8ms, 16ms, 32ms, 64ms,
100ms, 100ms, ... Cap at 100ms. Maximum overshoot
past deadline is one sleep interval (100ms).
