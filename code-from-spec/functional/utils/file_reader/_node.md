---
outputs:
  - id: file_reader
    path: code-from-spec/functional/utils/file_reader/output.md
---

# ROOT/functional/utils/file_reader

Sequential line reader for text files. Reads line by line
from the file — does not load the entire file into memory.
Normalizes line endings on read.

# Public

## Interface

```
record FileReader
  file_path: string

function OpenFileReader(file_path) -> FileReader
  errors:
    - file unreadable: the file cannot be opened.

function ReadLine(reader) -> line
  errors:
    - end of file: no more lines to read.

function SkipLines(reader, count)

function Close(reader)
```

`OpenFileReader` opens a file and prepares it for
sequential line-by-line reading. The file remains open
until `Close` is called.

`ReadLine` reads the next line from the file, normalizes
CRLF to LF, and returns the line without the terminator.
Raises "end of file" when there are no more lines.

`SkipLines` reads and discards `count` lines without
returning their content.

`Close` releases the file resource. After `Close`, any
call to `ReadLine` or `SkipLines` raises "end of file".

# Agent

Generate pseudocode for each function in the interface.

## Implementation guidance

- Read from the file stream sequentially — do not load
  the entire file into memory.
- The reader is forward-only. No seeking or rewinding.
- Memory usage must not depend on file size.
- A final line without a trailing newline is still a valid
  line. The "end of file" error is raised on the next call
  after the last line.
- `SkipLines` past the end of the file is not an error —
  subsequent `ReadLine` raises "end of file".
- The caller must call `Close` when done reading. Failing
  to close leaks the file handle.
