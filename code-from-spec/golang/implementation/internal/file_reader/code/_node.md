---
depends_on:
  - ARTIFACT/golang/interfaces/internal/file_reader(interface)
input: ARTIFACT/functional/logic/os/file_reader(file_reader)
outputs:
  - id: filereader
    path: internal/filereader/filereader.go
---

# ROOT/golang/implementation/internal/file_reader/code

Generates the filereader package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `bufio.Scanner` for line reading.
- Use `os.Open` for file opening.
- Normalize CRLF to LF before splitting lines.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
