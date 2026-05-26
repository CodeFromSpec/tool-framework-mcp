<!-- code-from-spec: ROOT/functional/utils/file_reader@KzQ9YfGoKF_mRWvwjUxxr1_Xzx0 -->

# FileReader

A forward-only, sequential, line-by-line file reader. The file is
read incrementally — memory usage does not scale with file size.


## Data Structures

```
record FileReader
  file_path: string   -- path of the opened file
  -- (internal stream state is implied but not exposed)
```


## Functions


### OpenFileReader(file_path) -> FileReader

Opens the file at `file_path` and prepares it for sequential
line-by-line reading. The file remains open until `Close` is called.

  1. Attempt to open the file at file_path for sequential reading.
     If the file cannot be opened (does not exist, permission denied,
     or any other I/O error), raise error "file unreadable".

  2. Create a FileReader record with:
       file_path = file_path
     Associate the open file stream with this reader internally.

  3. Return the FileReader record.

Errors:
  - "file unreadable": the file cannot be opened for any reason.


### ReadLine(reader) -> line

Reads the next line from the file stream and returns it without the
line terminator.

  1. If the reader has been closed, raise error "end of file".

  2. Attempt to read the next line from the file stream.
     If there are no more lines (the stream is exhausted),
     raise error "end of file".

  3. Normalize the line:
     Replace any trailing CRLF sequence ("\r\n") with LF ("\n"),
     then strip the trailing LF terminator from the line.
     If the last character is a bare CR ("\r"), strip it as well.

  4. Return the normalized line string (without any line terminator).

Notes:
  - A final line that has no trailing newline is still valid and
    is returned normally on its read call.
  - "end of file" is raised on the call *after* the last line has
    been returned, not on the call that returns the last line.

Errors:
  - "end of file": no more lines are available, or the reader is closed.


### SkipLines(reader, count)

Reads and discards `count` lines from the file stream without
returning their content.

  1. If the reader has been closed, return immediately (no error).

  2. Repeat `count` times:
       Attempt to read the next line from the file stream.
       If the stream is exhausted before `count` lines have been
       discarded, stop iterating — this is not an error.

  3. Return (no value).

Notes:
  - Skipping past the end of the file is permitted and does not
    raise an error. A subsequent ReadLine call will raise "end of file".

Errors:
  - (none)


### Close(reader)

Releases the file resource associated with `reader`.

  1. Release (close) the underlying file stream.
     Mark the reader as closed internally.

  2. After Close returns, any subsequent call to ReadLine or
     SkipLines on this reader must behave as if the stream is
     exhausted:
       - ReadLine raises "end of file".
       - SkipLines returns immediately without error.

  3. Calling Close on an already-closed reader is a no-op.

  4. Return (no value).

Errors:
  - (none)
