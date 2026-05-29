---
depends_on:
  - ROOT/functional/logic/os/path_utils
outputs:
  - id: file_reader
    path: code-from-spec/functional/logic/os/file_reader/output.md
---

# ROOT/functional/logic/os/file_reader

Sequential line reader for text files.

# Public

## Interface

```
record FileReader
  cfs_path: PathCfs

function FileOpen(cfs_path: PathCfs) -> FileReader
  errors:
    - (path errors): propagated from PathCfsToOs.
    - file unreadable: the path is valid but the file
      cannot be opened (does not exist, permission denied,
      or other OS error).

function FileReadLine(reader: FileReader) -> string
  errors:
    - end of file: no more lines to read.

function FileSkipLines(reader: FileReader, count: integer)

function FileClose(reader: FileReader)
```

`FileOpen` opens a file and prepares it for sequential
line-by-line reading, starting from the beginning of the
file. The caller must call `FileClose` when done — failing
to do so leaks the file handle.

`FileReadLine` reads the next line from the file,
normalizes CRLF to LF, and returns the line without the
terminator. Raises "end of file" when there are no more
lines.

`FileSkipLines` reads and discards `count` lines without
returning their content.

`FileClose` releases the file resource. After `FileClose`,
`FileReadLine` raises "end of file" and `FileSkipLines`
does nothing.

# Agent

Generate pseudocode for each function in the interface.

## Implementation guidance

- Convert `cfs_path` to an OS path internally using
  the path conversion from `path_utils`.
- Read from the file stream sequentially — do not load
  the entire file into memory.
- The reader is forward-only. No seeking or rewinding.
- Memory usage must not depend on file size.
- A final line without a trailing newline is still a valid
  line. The "end of file" error is raised on the next call
  after the last line.
- `FileSkipLines` past the end of the file is not an
  error — subsequent `FileReadLine` raises "end of file".

# Private

## Dependencies

`PathCfs` comes via `depends_on` from `path_utils`.
