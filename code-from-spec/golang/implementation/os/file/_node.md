---
depends_on:
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/os/file
output: internal/file/file.go
---

# SPEC/golang/implementation/os/file

# Agent

Implement the `file` package, including its interface.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.OpenFile` with appropriate flags for each mode:
  read = `O_RDONLY`, overwrite = `O_WRONLY|O_CREATE|O_TRUNC`,
  append = `O_WRONLY|O_CREATE|O_APPEND`.
- Normalize CRLF to LF before splitting lines.
- For file locking on Unix, use `syscall.Flock` with
  `LOCK_SH` (shared) or `LOCK_EX` (exclusive).
- For file locking on Windows, use `LockFileEx` via
  `golang.org/x/sys/windows`.
- Use `os.Rename` for `FileRename`.
- Use `os.Remove` for `FileDelete`.
- Create intermediate directories with `os.MkdirAll`.
