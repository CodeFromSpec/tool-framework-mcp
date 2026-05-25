<!-- code-from-spec: ROOT/functional/utils/file_reader@v1 -->

# FileReader

A forward-only, sequential line reader. CRLF is normalized to LF once at
open time so all consumers receive consistent LF-only content.

---

## Data structures

```
record FileReader
  file_path:    string          -- path of the file being read
  lines:        list of string  -- all lines after normalization and splitting
  current_index: integer        -- index of the next line to be returned (starts at 0)
```

---

## Functions

### OpenFileReader(file_path) -> FileReader

Opens the file at `file_path` and prepares it for sequential line-by-line
reading. Normalization and splitting happen once here.

1. Attempt to read the entire contents of the file at `file_path`.
   If the file cannot be opened or read, raise error "file unreadable".

2. Replace every occurrence of CRLF (`\r\n`) in the contents with LF (`\n`).
   This normalization is done once; all subsequent operations work on the
   normalized text.

3. Split the normalized contents on LF (`\n`) to produce a list of strings.
   - If the contents end with a trailing LF, the split produces an empty
     string as the last element — discard that trailing empty element so
     that a file ending with `\n` does not yield a phantom blank line.
   - A file that is completely empty produces an empty list.
   - A non-empty final line that has no trailing LF is included as-is.

4. Create a FileReader record with:
   - `file_path`    = the given file_path
   - `lines`        = the list produced in step 3
   - `current_index` = 0

5. Return the FileReader record.

---

### ReadLine(reader) -> line

Returns the next line from the reader without its line terminator, then
advances the reader by one position.

1. If `reader.current_index` >= length of `reader.lines`,
   raise error "end of file".

2. Let `line` = `reader.lines[reader.current_index]`.

3. Increment `reader.current_index` by 1.

4. Return `line`.

---

### SkipLines(reader, count)

Advances the reader forward by `count` lines without returning their content.
Skipping past the end of the file is not an error.

1. If `count` <= 0, do nothing and return.

2. Add `count` to `reader.current_index`.

3. If `reader.current_index` > length of `reader.lines`,
   set `reader.current_index` = length of `reader.lines`.
   (Clamp to end so subsequent ReadLine raises "end of file" cleanly.)

4. Return.

---

## Error conditions

| Error           | Raised by       | Condition                                              |
|-----------------|-----------------|--------------------------------------------------------|
| "file unreadable" | OpenFileReader | The file does not exist or cannot be opened for reading |
| "end of file"   | ReadLine        | No more lines remain (`current_index` >= line count)   |
