---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/os/file_reader
output: internal/filereader/filereader.go
---

# SPEC/golang/implementation/os/file_reader

# Agent

Implement the `filereader` package, including its interface.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.Open` for file opening.
- Normalize CRLF to LF before splitting lines.
