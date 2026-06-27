---
depends_on:
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/os/file
output: internal/file/file.go
---

# SPEC/golang/implementation/os/file/impl

Main implementation of the file operations package.

# Agent

Implement the `file` package, including its interface.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.OpenFile` with appropriate flags for each mode:
  read = `O_RDONLY`, overwrite = `O_WRONLY|O_CREATE|O_TRUNC`,
  append = `O_RDWR|O_CREATE|O_APPEND`.
  Append uses `O_RDWR` instead of `O_WRONLY` because on
  Windows, `O_APPEND` causes Go to replace `GENERIC_WRITE`
  with `FILE_APPEND_DATA`, which does not satisfy the
  `GENERIC_READ or GENERIC_WRITE` requirement of
  `LockFileEx`. `O_RDWR` provides `GENERIC_READ`, which
  satisfies the requirement on all platforms.
- Normalize CRLF to LF before splitting lines.
- After opening the file, call `fileLockShared(f)` for
  read mode or `fileLockExclusive(f)` for overwrite/append
  modes. These functions are defined in platform-specific
  files within the same package — do not implement them
  here.
- Use `os.Rename` for `FileRename`.
- Use `os.Remove` for `FileDelete`.
- Create intermediate directories with `os.MkdirAll`.
