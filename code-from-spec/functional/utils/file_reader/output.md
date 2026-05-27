<!-- code-from-spec: ROOT/functional/utils/file_reader@wQqRliwpgTtOs-7I43gWxY7Qzks -->

## Records

```
record FileReader
  file_path: string
  stream:    open file stream (internal, not exposed to caller)
  closed:    boolean
```

---

## Functions

### OpenFileReader(file_path) -> FileReader

1. Attempt to open the file at `file_path` for sequential reading.
   If the file cannot be opened, raise error "file unreadable".

2. Create a FileReader record with:
   - file_path set to `file_path`
   - stream set to the opened file stream
   - closed set to false

3. Return the FileReader record.

---

### ReadLine(reader) -> line

1. If `reader.closed` is true, raise error "end of file".

2. Attempt to read the next line from `reader.stream`.
   If no more data is available (stream is exhausted), raise error "end of file".

3. Normalize the line:
   - If the line ends with CRLF (`\r\n`), replace it with LF (`\n`).
   - Strip the trailing LF terminator from the line.

4. Return the normalized line string (without any line terminator).

---

### SkipLines(reader, count)

1. Repeat `count` times:
   a. Attempt to read and discard the next line from `reader.stream`.
      If `reader.closed` is true or the stream is exhausted,
      stop iterating immediately (do not raise an error).

---

### Close(reader)

1. If `reader.closed` is false:
   a. Release the file stream resource associated with `reader.stream`.
   b. Set `reader.closed` to true.
