<!-- code-from-spec: ROOT/functional/logic/os/file_reader@QbFKR2pL38o5iLfFNCd_3sgAyLg -->

# file_reader

## Records

```
record FileReader
  cfs_path: PathCfs
  os_path:  PathOs
  stream:   optional file stream
  closed:   boolean
```

## Functions

---

### FileOpen(cfs_path: PathCfs) -> FileReader

1. Call PathCfsToOs(cfs_path) to obtain an os_path.
   If PathCfsToOs raises any error, propagate it to the caller.

2. Attempt to open the file at os_path for reading.
   If the file cannot be opened (does not exist, permission
   denied, or other OS error), raise error "file unreadable".

3. Return a FileReader record with:
   - cfs_path set to cfs_path
   - os_path set to os_path
   - stream set to the open file stream
   - closed set to false

---

### FileReadLine(reader: FileReader) -> string

1. If reader.closed is true, raise error "end of file".

2. Read the next line from reader.stream up to and including
   the next newline character (LF), or until end of stream.
   If no characters are available (stream is at end of file),
   raise error "end of file".

3. Strip the trailing line terminator from the line:
   - If the line ends with CRLF (`\r\n`), remove both characters.
   - If the line ends with LF (`\n`), remove it.
   - If the line has no trailing newline (final line of file),
     leave it as-is.

4. Return the resulting string.

---

### FileSkipLines(reader: FileReader, count: integer)

1. If reader.closed is true, do nothing and return.

2. Repeat count times:
   - Call FileReadLine(reader).
   - If FileReadLine raises "end of file", stop iterating
     and return without error.

---

### FileClose(reader: FileReader)

1. If reader.closed is true, do nothing and return.

2. Close reader.stream, releasing the file handle.

3. Set reader.closed to true.
   Set reader.stream to absent (no value).
