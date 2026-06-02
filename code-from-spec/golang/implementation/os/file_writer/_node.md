---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_writer
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/os/file_writer
output: internal/filewriter/filewriter.go
---

# ROOT/golang/implementation/os/file_writer

# Agent

Implement the `filewriter` package, including its interface.

## Go-specific guidance

- Use `os.MkdirAll` for creating intermediate directories.
- Use `os.WriteFile` for writing file content.
- Write content as UTF-8 encoded text.
