# SPEC/golang/dependencies/golang-x-sys-windows

Extended Windows API bindings for Go:
`golang.org/x/sys/windows`.

# Public

## Import

```go
import "golang.org/x/sys/windows"
```

## Types

```go
type Handle uintptr
type Overlapped struct {
    Internal     uintptr
    InternalHigh uintptr
    Offset       uint32
    OffsetHigh   uint32
    HEvent       Handle
}
```

### Handle from *os.File

```go
h := windows.Handle(f.Fd())
```

## LockFileEx

```go
func LockFileEx(
    file Handle,
    flags uint32,
    reserved uint32,
    bytesLow uint32,
    bytesHigh uint32,
    ol *Overlapped,
) error
```

Locks a region of the specified file. To lock the
entire file, use `bytesLow = ^uint32(0)` and
`bytesHigh = ^uint32(0)`.

### Flags

| Constant | Type | Description |
|---|---|---|
| `LOCKFILE_FAIL_IMMEDIATELY` | `uint32` | Non-blocking; return immediately if lock unavailable |
| `LOCKFILE_EXCLUSIVE_LOCK` | `uint32` | Exclusive lock; omit for shared |

Combine with bitwise OR.

### Overlapped I/O

When called without `LOCKFILE_FAIL_IMMEDIATELY` and
the lock is not immediately available, `LockFileEx`
returns `ERROR_IO_PENDING` (`syscall.Errno`). The
caller must wait on the event in the `Overlapped`
struct.

```go
var ERROR_IO_PENDING syscall.Errno
```

Check with `errors.Is(err, windows.ERROR_IO_PENDING)`.

## CreateEvent

```go
func CreateEvent(
    sa *SecurityAttributes,
    manualReset uint32,
    initialState uint32,
    name *uint16,
) (Handle, error)
```

Creates an event object. For lock timeout support:

```go
event, err := windows.CreateEvent(nil, 1, 0, nil)
if err != nil {
    return err
}
defer windows.CloseHandle(event)
```

Set `Overlapped.HEvent = event` before calling
`LockFileEx`.

## WaitForSingleObject

```go
func WaitForSingleObject(handle Handle, waitMilliseconds uint32) (uint32, error)
```

Waits for the specified object to be signaled or for
the timeout to elapse. **Returns `(uint32, error)`** —
the first value is the wait result, not the error.

### Wait result constants

| Constant | Underlying type | Description |
|---|---|---|
| `WAIT_OBJECT_0` | `syscall.Errno` | Object was signaled (lock acquired) |
| `WAIT_TIMEOUT` | `syscall.Errno` | Timeout elapsed |

**Type mismatch**: the result is `uint32` but the
constants are `syscall.Errno`. Compare with explicit
conversion:

```go
result, err := windows.WaitForSingleObject(event, timeoutMs)
if result == uint32(windows.WAIT_OBJECT_0) {
    // lock acquired
}
if result == uint32(windows.WAIT_TIMEOUT) {
    // timeout
}
```

## CancelIo

```go
func CancelIo(s Handle) error
```

Cancels pending I/O operations issued by the calling
thread on the specified handle. Call after a wait
timeout to cancel the pending `LockFileEx`.

## CloseHandle

```go
func CloseHandle(handle Handle) error
```

Closes an open object handle. Use to clean up event
handles created with `CreateEvent`.
