<!-- code-from-spec: ROOT/functional/utils/file_reader@Wx-26RPkivGtDg3vtpykGnkT3a8 -->

# FileReader

A forward-only, sequential line reader. CRLF is normalized to LF once at
open time. Callers always receive LF-only content regardless of the
original line endings in the file.

---

## Data structures

```
record FileReader
  file_path : string          -- path of the opened file
  lines     : list of strings -- all lines after normalization and splitting
  position  : integer         -- index of the next line to return (0-based)
```

---

## Functions

### OpenFileReader(file_path) -> FileReader

Opens the file at `file_path` and prepares it for sequential reading.

1. Attempt to read the entire contents of the file at `file_path`.
   If the file cannot be opened or read, raise error "file unreadable".

2. Replace every CRLF sequence (`\r\n`) in the contents with LF (`\n`).
   This normalization happens exactly once, here, so all downstream
   operations work only with LF.

3. Split the normalized contents on LF (`\n`) to produce a list of lines.
   Do NOT include the LF character in any line.

   Special case — trailing newline:
   If the last character of the normalized content is LF, the split
   produces an empty string as the final element. Remove that trailing
   empty element so that a file ending with a newline does not yield a
   phantom empty line.
   A file whose last line has NO trailing newline is left as-is; that
   final non-empty line is a valid line and must be returned.

4. Create a FileReader record:
   - file_path = <file_path>
   - lines     = <the list produced in step 3>
   - position  = 0

5. Return the FileReader record.

---

### ReadLine(reader) -> line

Returns the next line and advances the reader by one position.

1. If `reader.position` is greater than or equal to the length of
   `reader.lines`, raise error "end of file".

2. Let `line` = `reader.lines[reader.position]`.

3. Increment `reader.position` by 1.

4. Return `line`.
   The returned string never contains a line terminator.

---

### SkipLines(reader, count)

Advances the reader forward by `count` lines without returning their
content. Skipping past the end of the file is not an error.

1. Add `count` to `reader.position`.

2. If `reader.position` now exceeds the length of `reader.lines`,
   clamp `reader.position` to the length of `reader.lines`.
   This ensures subsequent calls to `ReadLine` raise "end of file"
   immediately rather than going out of bounds.

3. Return nothing.

---

## Error conditions

| Error          | Raised by      | Condition                                              |
|----------------|----------------|--------------------------------------------------------|
| "file unreadable" | OpenFileReader | The file at `file_path` cannot be opened or read.  |
| "end of file"     | ReadLine       | `position` >= length of `lines` (no more lines).   |

---

## Invariants

- CRLF normalization happens once in `OpenFileReader`. No other function
  performs normalization.
- The reader is strictly forward-only. Neither `ReadLine` nor `SkipLines`
  ever decrements `position`.
- `SkipLines` with `count = 0` is a no-op.
- `SkipLines` past the end of the file leaves `position` equal to the
  length of `lines`, so the next `ReadLine` raises "end of file".
