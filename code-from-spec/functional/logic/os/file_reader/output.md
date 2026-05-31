<!-- code-from-spec: ROOT/functional/logic/os/file_reader@0QASvP5yf90CrNLYW_O8jECYgTc -->

## Records

```
record FileReader
  cfs_path: PathCfs
  internal_stream: file stream handle (opaque, managed by OS layer)
  is_closed: boolean
```

---

## Functions

### FileOpen(cfs_path: PathCfs) -> FileReader

1. Call PathCfsToOs(cfs_path) to obtain an OS-native absolute path.
   If PathCfsToOs raises any PathUtils error, propagate it unchanged.

2. Attempt to open the file at the resulting OS path for sequential
   reading, starting at the beginning.
   If the file cannot be opened (does not exist, permission denied,
   or any other OS error), raise error "FileUnreadable".

3. Create and return a FileReader record with:
   - cfs_path set to the provided cfs_path
   - internal_stream set to the opened file stream handle
   - is_closed set to false

---

### FileReadLine(reader: FileReader) -> string

1. If reader.is_closed is true, raise error "EndOfFile".

2. Attempt to read the next line from reader.internal_stream.
   If there are no more bytes to read (end of stream reached),
   raise error "EndOfFile".

3. Collect bytes until a newline character (LF, `\n`) is encountered
   or the end of the stream is reached.
   Do not load the entire file into memory — read incrementally.

4. Normalize the collected line:
   If the line ends with a carriage return (CR, `\r`), strip it.
   This converts CRLF line endings to LF.

5. Return the line content without the terminating newline character.

---

### FileSkipLines(reader: FileReader, count: integer)

1. If reader.is_closed is true, do nothing and return.

2. Repeat count times:
     Call FileReadLine(reader).
     If FileReadLine raises "EndOfFile", stop immediately and return.
     Otherwise, discard the returned string and continue.

---

### FileClose(reader: FileReader)

1. If reader.is_closed is true, do nothing and return.

2. Release reader.internal_stream, freeing the underlying OS file handle.

3. Set reader.is_closed to true.
```
