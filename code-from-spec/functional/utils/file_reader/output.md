<!-- code-from-spec: ROOT/functional/utils/file_reader@G6f-O2lsvyS72vgeLxuFq8_1-fE -->

# FileReader — Pseudocode

## Data Structures

```
record FileReader
  file_path: string        -- path of the file being read
  stream:    file stream   -- open sequential stream positioned at the next unread byte
  exhausted: boolean       -- true once the underlying stream has no more data
```

---

## function OpenFileReader(file_path) -> FileReader

Opens a file and returns a FileReader positioned before the first line.

1. Attempt to open the file at `file_path` for sequential reading.
   If the file cannot be opened for any reason, raise error "file unreadable".

2. Create a new FileReader record:
   - file_path ← file_path
   - stream    ← the open file stream
   - exhausted ← false

3. Return the FileReader record.

---

## function ReadLine(reader) -> line

Reads the next line from `reader`, normalizes CRLF to LF, and returns
the line text without the line terminator.

1. If reader.exhausted is true, raise error "end of file".

2. Attempt to read bytes from reader.stream up to and including the
   next LF character (U+000A), or until the stream is depleted,
   whichever comes first.

   If no bytes are available at all (stream was already at end):
     Set reader.exhausted ← true.
     Raise error "end of file".

3. Let `raw_line` be the bytes collected in step 2, decoded as text.

4. Normalize CRLF:
   If `raw_line` ends with CR LF (i.e., "\r\n"), remove both characters.
   Else if `raw_line` ends with LF ("\n"), remove that character.
   Else if `raw_line` ends with CR ("\r"), remove that character.
   -- A final line with no terminator is left as-is (no characters removed).

5. If the stream reached its end during step 2 (no further bytes remain):
   Set reader.exhausted ← true.
   -- Do NOT raise "end of file" now; the final partial line is valid.
   -- "end of file" will be raised on the NEXT call to ReadLine.

6. Return the normalized line text (without any terminator).

---

## function SkipLines(reader, count)

Reads and discards `count` lines from `reader`. Silently stops if
the end of file is reached before `count` lines are consumed.

1. Let `remaining` ← count.

2. While remaining > 0:
   a. If reader.exhausted is true, stop (return without error).
   b. Call ReadLine(reader).
      If it raises "end of file", stop (return without error).
      Otherwise, discard the returned line.
   c. Decrement remaining by 1.

3. Return. (No return value; no error on hitting end of file.)

---

## Error Conditions

| Error          | Raised by                         | Meaning                                         |
|----------------|-----------------------------------|-------------------------------------------------|
| "file unreadable" | OpenFileReader                 | The file at `file_path` could not be opened.   |
| "end of file"  | ReadLine                          | No more lines remain in the file stream.        |

`SkipLines` never raises an error — reading past the end of the file
is silently absorbed, and subsequent calls to `ReadLine` will raise
"end of file" as normal.

---

## Contracts and Invariants

- The reader is forward-only. No seeking or rewinding is supported.
- The file stream is opened once in `OpenFileReader` and read
  incrementally; the entire file is never loaded into memory.
- Memory usage is bounded by the length of a single line, not
  the total file size.
- CRLF normalization happens per line as it is read, not in bulk.
- A file whose last line has no trailing newline is fully supported;
  that line is returned normally, and the "end of file" error is
  deferred to the call after it.
