<!-- code-from-spec: ROOT/functional/logic/os/file_reader@QbFKR2pL38o5iLfFNCd_3sgAyLg -->

## FileReader

A record representing an open file prepared for sequential line-by-line reading.

```
record FileReader
  cfs_path: PathCfs
  os_path:  PathOs
  stream:   file stream handle (optional — absent when closed)
  buffer:   line buffer for the stream
```

---

## function FileOpen(cfs_path: PathCfs) -> FileReader

  1. Call PathCfsToOs(cfs_path) to obtain the OS-native absolute path.
     If PathCfsToOs raises any error, propagate it unchanged.

  2. Open the file at the resulting OS path for sequential reading.
     If the file cannot be opened for any reason (does not exist,
     permission denied, or other OS error), raise error "file unreadable".

  3. Initialize a line buffer attached to the open stream.

  4. Return a FileReader record with:
     - cfs_path set to the provided cfs_path
     - os_path set to the converted OS path
     - stream set to the open file stream
     - buffer set to the initialized line buffer

---

## function FileReadLine(reader: FileReader) -> string

  1. If reader.stream is absent (the reader has been closed),
     raise error "end of file".

  2. Read the next line from reader.buffer.
     If no more data is available, raise error "end of file".

  3. Strip the trailing line terminator from the line:
     - If the line ends with CRLF ("\r\n"), remove both characters.
     - If the line ends with LF ("\n"), remove it.
     - If the line ends with CR ("\r"), remove it.
     A final line without any trailing terminator is still a valid line.

  4. Return the resulting string.

---

## function FileSkipLines(reader: FileReader, count: integer)

  1. If reader.stream is absent (the reader has been closed), do nothing and return.

  2. Repeat count times:
     a. Read the next line from reader.buffer.
        If no more data is available (end of file), stop iterating immediately.
        Do not raise an error.

  3. Return. (No value is returned.)

---

## function FileClose(reader: FileReader)

  1. If reader.stream is absent (already closed), do nothing and return.

  2. Release the file stream resource associated with reader.stream.

  3. Set reader.stream to absent.

  4. Return. (No value is returned.)
