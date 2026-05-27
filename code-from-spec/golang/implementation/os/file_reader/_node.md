---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
input: ARTIFACT/functional/logic/os/file_reader(file_reader)
outputs:
  - id: filereader
    path: internal/filereader/filereader.go
---

# ROOT/golang/implementation/os/file_reader

# Agent

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.Open` for file opening.
- Normalize CRLF to LF before splitting lines.
