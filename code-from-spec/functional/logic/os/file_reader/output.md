<!-- code-from-spec: ROOT/functional/logic/os/file_reader@USyaoOilQaTtIveqaq_wLguG5xg -->

# FileReader

## Records

```
record FileReader
  cfs_path: PathCfs
  os_path:  PathOs
  stream:   optional file stream handle
  closed:   boolean
```

## Functions

---

### FileOpen(cfs_path: PathCfs) -> FileReader

1. Call PathCfsToOs with cfs_path.
   If PathCfsToOs raises an error, propagate it unchanged.

2. Open a file stream at the resolved OS path for sequential reading.
   If the file cannot be opened, raise error "file unreadable".

3. Return a FileReader record with:
   - cfs_path set to the given cfs_path
   - os_path set to the resolved OS path
   - stream set to the opened file stream
   - closed set to false

---

### FileReadLine(reader: FileReader) -> string

1. If reader.closed is true, raise error "end of file".

2. Read the next line from reader.stream.
   If there are no more bytes to read, raise error "end of file".

3. Strip the trailing line terminator from the line:
   - If the line ends with CRLF ("\r\n"), remove both characters.
   - Else if the line ends with LF ("\n"), remove it.
   - Else if the line ends with CR ("\r"), remove it.
   (A final line with no trailing terminator is returned as-is.)

4. Return the resulting string.

---

### FileSkipLines(reader: FileReader, count: integer)

1. If reader.closed is true, do nothing and return.

2. Repeat count times:
   a. Read the next line from reader.stream.
      If there are no more bytes to read, stop iterating.
      (Reaching end of file during a skip is not an error.)

---

### FileClose(reader: FileReader)

1. If reader.closed is true, do nothing and return.

2. Close reader.stream, releasing the file handle.

3. Set reader.closed to true.
   Set reader.stream to absent.
```
