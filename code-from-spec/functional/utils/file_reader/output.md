<!-- code-from-spec: ROOT/functional/utils/file_reader@ktDtBxYIuDy9X7w6dvy9zbfpFRg -->

# FileReader

A utility for forward-only, sequential, line-by-line reading of a file.
Memory usage does not depend on file size — lines are read from the file
stream one at a time, never loaded entirely into memory.

---

## Data Structures

```
record FileReader
  file_path: string       -- path to the file being read
  stream:    file stream  -- open file stream positioned at the current read offset
```

---

## Functions

### OpenFileReader(file_path) -> FileReader

Opens the file at `file_path` and returns a FileReader ready for
sequential reading from the beginning of the file.

  1. Attempt to open the file at `file_path` for reading.
     If the file cannot be opened (does not exist, permission denied,
     or any other I/O error), raise error "file unreadable".

  2. Create a FileReader record with:
     - file_path set to `file_path`
     - stream set to the newly opened file stream, positioned at the
       start of the file

  3. Return the FileReader record.

---

### ReadLine(reader) -> line

Reads the next line from the reader's file stream, normalizes line
endings, and returns the line text without the terminator.

  1. Attempt to read the next line from `reader.stream`.
     - A line ends at the next LF (`\n`), CRLF (`\r\n`), or end of file.
     - If the stream is already at end of file (no bytes remain),
       raise error "end of file".

  2. Strip the trailing line terminator from the raw line:
     - If the line ends with CRLF (`\r\n`), remove both characters.
     - If the line ends with LF (`\n`) only, remove that character.
     - If the line has no trailing terminator (final line of file without
       a newline), leave the content as-is. This is still a valid line.

  3. Return the resulting string (the line content without terminator).

Error conditions:
  - "end of file": raised when there are no more lines to read. This is
    raised on the call *after* the last line has been returned, not during
    the call that returns the last line.

---

### SkipLines(reader, count)

Reads and discards `count` lines from the reader's file stream without
returning their content. Skipping past the end of the file is not an error.

  1. If `count` is less than or equal to 0, do nothing and return.

  2. Repeat `count` times:
     a. Attempt to read the next line from `reader.stream` (same
        mechanics as ReadLine, including CRLF normalization — though
        the content is discarded).
     b. If "end of file" is reached before all `count` lines are consumed,
        stop immediately without raising an error. The reader is now
        positioned at end of file.

  3. Return. (No value is returned.)

---

## Contracts and Invariants

- **Forward-only**: the reader does not support seeking or rewinding.
  Lines may only be read in the order they appear in the file.
- **Sequential streaming**: the underlying file is read incrementally.
  The entire file is never loaded into memory at once.
- **CRLF normalization**: every line returned by `ReadLine` uses LF
  line endings internally. Callers never observe `\r` at end of a line.
- **End-of-file boundary**: the "end of file" error is raised on the
  first `ReadLine` call after the last line has been successfully returned.
  A file with N lines yields exactly N successful `ReadLine` results before
  the error is raised.
- **SkipLines past EOF**: calling `SkipLines` when fewer than `count`
  lines remain is safe. The reader silently reaches end of file, and
  a subsequent `ReadLine` raises "end of file" as normal.
