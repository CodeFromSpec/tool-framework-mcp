# SPEC/golang/implementation/os/file

Go implementation of the file operations component,
split into three artifacts: the main implementation
and two platform-specific locking files.

# Public

## Locking interface

The main implementation calls two unexported functions
for file locking. These are defined in platform-specific
files within the same package:

```go
func fileLockShared(f *os.File, timeoutMs int) error
func fileLockExclusive(f *os.File, timeoutMs int) error
```

Both functions attempt to acquire the lock within the
given timeout. If `timeoutMs` is zero, attempt
non-blocking (fail immediately if lock is not
available). If the lock cannot be acquired within the
timeout, return `ErrLockTimeout`. They operate on the
file descriptor of the given `*os.File`.
