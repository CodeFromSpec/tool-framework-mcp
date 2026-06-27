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
func fileLockShared(f *os.File) error
func fileLockExclusive(f *os.File) error
```

Both functions block until the lock is acquired. They
operate on the file descriptor of the given `*os.File`.
Errors are returned if the lock cannot be acquired.
