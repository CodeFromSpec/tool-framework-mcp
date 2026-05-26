<!-- code-from-spec: ROOT/functional/utils/file_reader@G6f-O2lsvyS72vgeLxuFq8_1-fE -->

# file_reader — Pseudocode

## Records

```
record FileReader
  file_path: string
  stream:    open file stream positioned at the first byte
```

---

## Functions

### OpenFileReader(file_path) -> FileReader

Opens a file for sequential, line-by-line reading. The file
stream remains open until all lines are consumed or the reader
is discarded. The file is NOT loaded into memory in full.

Parameters:
  file_path  string  — path to the file to open

Returns:
  FileReader record with the stream positioned at the first byte

Errors:
  "file unreadable" — the file cannot be opened (missing,
                      permission denied, or any other I/O error)

Steps:

  1. Attempt to open the file at file_path for reading.
     If the file cannot be opened, raise error "file unreadable".

  2. Create a FileReader record:
       file_path = file_path
       stream    = the open file stream

  3. Return the FileReader record.

---

### ReadLine(reader) -> line

Reads the next line from the file stream, normalizes CRLF to LF,
and returns the line without its terminator.

Parameters:
  reader  FileReader  — an open reader returned by OpenFileReader

Returns:
  string — the next line, with no trailing newline character

Errors:
  "end of file" — there are no more lines to read

Steps:

  1. Read bytes from reader.stream up to and including the next
     LF character, or until the stream is exhausted.

     If no bytes were read and the stream is exhausted,
     raise error "end of file".

  2. Collect the bytes read as a raw line string (which may end
     with LF, CRLF, or nothing if it is the final line without
     a trailing newline).

  3. If the raw line ends with CRLF, remove the trailing CR and LF.
     Else if the raw line ends with LF, remove the trailing LF.
     Otherwise (final line without newline), keep the line as-is.

  4. Return the resulting string.

---

### SkipLines(reader, count)

Reads and discards `count` lines from the reader without returning
their content. Reaching the end of file during skipping is not an
error — subsequent calls to ReadLine will raise "end of file".

Parameters:
  reader  FileReader  — an open reader returned by OpenFileReader
  count   integer     — number of lines to skip (must be >= 0)

Returns:
  nothing

Errors:
  (none — exhausting the file is silently ignored)

Steps:

  1. Set remaining = count.

  2. While remaining > 0:
       Attempt to call ReadLine(reader).
       If ReadLine raises "end of file", stop immediately.
       Otherwise, discard the returned line.
       Decrement remaining by 1.

  3. Return (no value).
```
