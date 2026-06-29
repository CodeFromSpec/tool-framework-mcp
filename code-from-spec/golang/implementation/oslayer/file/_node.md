# SPEC/golang/implementation/oslayer/file

File operations with automatic locking, split into
three artifacts: the main implementation and two
platform-specific locking files.

# Public

## Locking interface

Functions for platform-specific file locking.

```go
func fileLockShared(f *os.File, timeoutMs int) error
func fileLockExclusive(f *os.File, timeoutMs int) error
```

#### fileLockShared

Acquires a shared lock on the file descriptor of `f`.

#### fileLockExclusive

Acquires an exclusive lock on the file descriptor of `f`.

#### Timeout semantics

If `timeoutMs` is zero or negative, attempt
non-blocking. If the lock cannot be acquired
immediately, return `ErrLockTimeout`.

If `timeoutMs` is positive, retry until the lock is
acquired or the timeout expires. If the timeout
expires, return `ErrLockTimeout`.

#### Errors

- `ErrLockTimeout`: lock not acquired within the
  timeout.
- `ErrLockFailed`: the lock operation failed for
  reasons other than timeout (invalid file descriptor,
  I/O error, or other OS-level failure).
