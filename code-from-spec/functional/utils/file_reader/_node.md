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
```

`OpenFileReader` opens a file and prepares it for
sequential line-by-line reading. The file remains open
until all lines are consumed or the reader is discarded.

`ReadLine` reads the next line from the file, normalizes
CRLF to LF, and returns the line without the terminator.
Raises "end of file" when there are no more lines.

`SkipLines` reads and discards `count` lines without
returning their content.

# Agent

## Behavior

- Each `ReadLine` call reads from the file stream — the
  file is not loaded entirely into memory.
- CRLF sequences are normalized to LF as each line is
  read. The line terminator is not included in the
  returned string.
- A final line without a trailing newline is still a valid
  line and is returned normally. The "end of file" error
  is raised on the next call after the last line.
- `SkipLines` past the end of the file is not an error —
  subsequent `ReadLine` raises "end of file".

## Contracts

- The reader is forward-only. No seeking or rewinding.
- The file is read sequentially — memory usage does not
  depend on file size.
