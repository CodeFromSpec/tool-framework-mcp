<!-- code-from-spec: ROOT/functional/logic/os/file_reader@WSCzbRJHABsyr4ovlmyycOuEIpI -->

## FileReader

A record representing an open file positioned for sequential line-by-line reading.

```
record FileReader
  cfs_path: CfsPath
  os_path:  PathOs
  stream:   file stream handle (OS-level)
  closed:   boolean
```

---

### function FileOpen(cfs_path: CfsPath) -> FileReader

  1. Call PathCfsToOs(cfs_path) to obtain an OS path.
     If PathCfsToOs raises a path error, propagate it unchanged.

  2. Open the file at the resolved OS path for sequential reading
     from the beginning.
     If the file cannot be opened, raise error "file unreadable".

  3. Return a FileReader record with:
     - cfs_path set to the provided cfs_path
     - os_path set to the resolved OS path
     - stream set to the opened file stream handle
     - closed set to false

---

### function FileReadLine(reader: FileReader) -> string

  1. If reader.closed is true, raise error "end of file".

  2. Attempt to read the next line from reader.stream.
     If there are no more lines (stream is at end of file),
     raise error "end of file".

  3. Strip the line terminator from the end of the line:
     - If the line ends with CRLF ("\r\n"), remove both characters.
     - If the line ends with LF ("\n"), remove it.
     - If the line has no terminator (final line of file),
       use it as-is.

  4. Return the resulting string.

---

### function FileSkipLines(reader: FileReader, count: integer)

  1. If reader.closed is true, return immediately.

  2. Repeat count times:
     Attempt to read and discard the next line from reader.stream.
     If the stream reaches end of file before count lines are
     consumed, stop iterating — do not raise an error.

---

### function FileClose(reader: FileReader)

  1. If reader.closed is true, return immediately.

  2. Release reader.stream (close the OS file handle).

  3. Set reader.closed to true.
