# SPEC/golang/dependencies/unix-syscall-flock

Go standard library `syscall` package: file locking
primitives for Unix systems via `flock(2)`.

# Public

## Import

```go
import "syscall"
```

## Flock

```go
func Flock(fd int, how int) error
```

Applies or removes an advisory lock on the open file
referred to by `fd`. The `how` argument determines the
operation:

| Constant | Type | Value | Description |
|---|---|---|---|
| `syscall.LOCK_SH` | `int` | 1 | Shared lock |
| `syscall.LOCK_EX` | `int` | 2 | Exclusive lock |
| `syscall.LOCK_NB` | `int` | 4 | Non-blocking (OR with SH or EX) |

Combine flags with bitwise OR:
`syscall.LOCK_SH | syscall.LOCK_NB` for a non-blocking
shared lock.

### Return values

Returns `nil` on success. On failure, returns an
`error` wrapping a `syscall.Errno` value.

### EWOULDBLOCK

```go
var EWOULDBLOCK syscall.Errno
```

`syscall.EWOULDBLOCK` is a `syscall.Errno` value
(type `uintptr`). Returned by `Flock` when
`LOCK_NB` is set and the lock cannot be acquired
immediately.

Check with `errors.Is`:

```go
err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
if errors.Is(err, syscall.EWOULDBLOCK) {
    // lock held by another process
}
```

## File descriptor from *os.File

```go
fd := int(f.Fd())
```

`(*os.File).Fd()` returns `uintptr`. Cast to `int`
for `Flock`.
