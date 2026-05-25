---
outputs:
  - id: file_reader
    path: code-from-spec/functional/utils/file_reader/output.md
---

# ROOT/functional/utils/file_reader

Sequential line reader for text files. Normalizes line
endings on read.

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
sequential line-by-line reading.

`ReadLine` returns the next line without the line
terminator. CRLF is normalized to LF before splitting.
Raises "end of file" when there are no more lines.

`SkipLines` advances the reader by `count` lines without
returning their content.

# Agent

## Behavior

- Lines are split on LF (after CRLF → LF normalization).
- Line terminators are not included in the returned line.
- A final line without a trailing newline is still a valid
  line and is returned normally. The "end of file" error
  is raised on the next call after the last line.
- `SkipLines` past the end of the file is not an error —
  subsequent `ReadLine` raises "end of file".

## Contracts

- Normalization happens once at read time — all consumers
  receive LF-only content.
- The reader is forward-only. No seeking or rewinding.
