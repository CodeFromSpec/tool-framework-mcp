---
depends_on:
  - ARTIFACT/golang/internal/file_reader/interface(interface)
input: ARTIFACT/golang/internal/file_reader/code(filereader)
outputs:
  - id: filereader_test
    path: internal/filereader/filereader_test.go
---

# ROOT/golang/internal/file_reader/tests

Test cases for the filereader package.

# Agent

## Context

Each test uses `t.TempDir()` to create an isolated
temporary directory. Test files are created with controlled
content. Use table-driven tests where appropriate.

## Happy Path

### Opens and reads all lines

Create a file with multiple lines (LF endings). Call
`OpenFileReader`, then `ReadLine` repeatedly. Expect each
line returned without terminator. After the last line,
`ReadLine` returns `ErrEndOfFile`.

### Normalizes CRLF to LF

Create a file with CRLF line endings. Expect `ReadLine`
returns lines without any CR or LF characters.

### Reads file with no trailing newline

Create a file where the last line has no trailing newline.
Expect the last line is returned normally, and the next
`ReadLine` returns `ErrEndOfFile`.

### SkipLines advances the reader

Create a file with 5 lines. Call `SkipLines(2)`, then
`ReadLine`. Expect the third line is returned.

### SkipLines past end of file

Create a file with 2 lines. Call `SkipLines(10)`. Expect
no error. Then `ReadLine` returns `ErrEndOfFile`.

## Edge Cases

### Empty file

Create an empty file. Call `OpenFileReader`, then
`ReadLine`. Expect `ErrEndOfFile` immediately.

### Single line without newline

Create a file containing just `"hello"` with no newline.
Expect `ReadLine` returns `"hello"`, then `ErrEndOfFile`.

## Failure Cases

### File does not exist

Call `OpenFileReader` with a non-existent path.
Expect `errors.Is(err, ErrOpen)`.
