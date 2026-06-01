---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/logic/os/file_reader(file_reader)
outputs:
  - id: filereader
    path: internal/filereader/filereader.go
---

# ROOT/golang/implementation/os/file_reader

# Agent

Implement the `filereader` package, including its interface.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.Open` for file opening.
- Normalize CRLF to LF before splitting lines.
