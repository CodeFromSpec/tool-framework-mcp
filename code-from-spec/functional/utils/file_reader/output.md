<!-- code-from-spec: ROOT/functional/utils/file_reader@REqX_M4VNvZwSFIKf6wjF_yqvhM -->

## FileReader

A record representing an open file and its current read position.

Fields:
- file_path: string — the path of the opened file
- stream: open file stream — the underlying sequential file handle
- closed: boolean — whether Close has been called

---

function OpenFileReader(file_path) -> FileReader

  1. Attempt to open the file at file_path for sequential reading.
     If the file cannot be opened, raise error "file unreadable".

  2. Create a FileReader record with:
     - file_path set to file_path
     - stream set to the opened file handle
     - closed set to false

  3. Return the FileReader record.

---

function ReadLine(reader) -> line

  1. If reader.closed is true, raise error "end of file".

  2. Attempt to read the next line from reader.stream.
     If there are no more bytes to read, raise error "end of file".

  3. Collect bytes until a LF character ("\n") is encountered,
     end of file is reached, or the stream is exhausted.
     Do not load the entire file into memory — read one character
     at a time or use a fixed-size buffer.

  4. Strip the trailing LF from the collected line, if present.

  5. If the line ends with CR ("\r"), strip that CR as well.
     This normalizes CRLF line endings to plain text.

  6. Return the resulting line string (without any line terminator).

---

function SkipLines(reader, count)

  1. If reader.closed is true, return immediately without error.

  2. Repeat count times:
     a. Attempt to read the next line from reader.stream
        (consuming it and discarding its content).
        If end of file is reached, stop iterating — this is not an error.

  3. Return. No value is produced.

---

function Close(reader)

  1. If reader.closed is true, return immediately (closing is idempotent).

  2. Release reader.stream, freeing the file handle.

  3. Set reader.closed to true.

  4. Return. No value is produced.
