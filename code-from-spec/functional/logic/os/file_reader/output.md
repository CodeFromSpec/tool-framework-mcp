<!-- code-from-spec: ROOT/functional/logic/os/file_reader@-WxkQfy7t8xsASdybsJ1_LbaxBU -->

# File Reader

## Records

```
record FileReader
  cfs_path: pathutils.PathCfs
  os_path:  pathutils.PathOs
  stream:   open file stream (internal, line-buffered)
  closed:   boolean
```

## Functions

---

### FileOpen(cfs_path: pathutils.PathCfs) -> FileReader

  1. Call PathCfsToOs(cfs_path) to obtain an OS-native absolute path.
     If PathCfsToOs raises any PathUtils error, propagate it unchanged.

  2. Open the file at the resulting OS path for sequential reading.
     If the file cannot be opened (does not exist, permission denied,
     or any other OS error), raise error "FileUnreadable".

  3. Create a FileReader record with:
     - cfs_path set to the given cfs_path
     - os_path  set to the resolved OS path
     - stream   set to the open file stream, positioned at the start
     - closed   set to false

  4. Return the FileReader record.

---

### FileReadLine(reader: FileReader) -> string

  1. If reader.closed is true, raise error "EndOfFile".

  2. Read the next line from reader.stream (up to and including the
     next newline character, or end of stream).
     If no bytes are available (end of stream has been reached and no
     unterminated content remains), raise error "EndOfFile".

  3. Strip the trailing line terminator from the line:
     - If the line ends with CRLF ("\r\n"), remove both characters.
     - Else if the line ends with LF ("\n"), remove it.
     - Else (final line with no trailing newline), leave as-is.

  4. Return the resulting string.

---

### FileSkipLines(reader: FileReader, count: integer)

  1. If reader.closed is true, do nothing and return.

  2. Repeat count times:
       Call FileReadLine(reader).
       If FileReadLine raises "EndOfFile", stop iterating and return.

  3. Return (no error is raised when end of file is reached early).

---

### FileClose(reader: FileReader)

  1. If reader.closed is true, do nothing and return.

  2. Release reader.stream (close the underlying OS file handle).

  3. Set reader.closed to true.

  4. Return.
