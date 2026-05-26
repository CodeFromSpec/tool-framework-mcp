<!-- code-from-spec: ROOT/functional/utils/file_reader@wQqRliwpgTtOs-7I43gWxY7Qzks -->

# FileReader

A forward-only, sequential line-by-line file reader. The file is
read incrementally — memory usage does not depend on file size.
CRLF line endings are normalized to LF on every read.

---

## Data Structures

```
record FileReader
  file_path: string    -- path of the file being read
  stream:    handle    -- open file stream handle
  closed:    boolean   -- true after Close is called
```

---

## Functions

### OpenFileReader(file_path) -> FileReader

Opens the file at `file_path` and returns a FileReader ready for
sequential reading. The file stays open until `Close` is called.

1. Attempt to open the file at `file_path` for sequential reading.
   If the file cannot be opened (not found, permission denied, etc.),
   raise error "file unreadable".

2. Create a FileReader record:
   - file_path = file_path
   - stream    = the open file handle
   - closed    = false

3. Return the FileReader record.

---

### ReadLine(reader) -> line

Reads and returns the next line from the file, without its line
terminator. Raises "end of file" when there are no more lines.

1. If reader.closed is true, raise error "end of file".

2. Attempt to read the next line from reader.stream.
   If the stream is at end of file (no more data), raise error "end of file".

3. Strip the trailing line terminator from the raw line:
   - If the raw line ends with CRLF (`\r\n`), remove both characters.
   - Else if the raw line ends with LF (`\n`), remove that character.
   - Else leave the line unchanged (final line with no trailing newline
     is valid and returned as-is).

4. Return the resulting line string.

---

### SkipLines(reader, count)

Reads and discards `count` lines without returning their content.
Reaching end of file during skipping is not an error.

1. If reader.closed is true, do nothing and return.

2. Repeat `count` times:
   a. Call ReadLine(reader).
      If ReadLine raises "end of file", stop iterating and return.

3. Return (no value).

---

### Close(reader)

Releases the file handle. After Close, any subsequent call to
ReadLine or SkipLines raises "end of file".

1. If reader.closed is already true, do nothing and return.

2. Close reader.stream, releasing the underlying file resource.

3. Set reader.closed = true.

4. Return (no value).
