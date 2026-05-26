<!-- code-from-spec: ROOT/functional/mcp_tools/hash_fragment@WjzWuEiqvzeDqY2Nd6_I3I__3aQ -->

# HashFragment

Computes a SHA-1 / base64url hash of a specific line range within a file.
The caller provides a relative file path and a line range; the tool returns
a 27-character base64url-encoded (no padding) digest of those lines joined
with LF characters.

---

## Dependencies

- **PathValidation** — validates that the path is safe (non-empty, relative,
  no directory traversal, resolves inside the project root).
- **FileReader** — opens the file and provides sequential line-by-line access
  with normalized line endings (CRLF → LF).

---

## Data structures

```
record HashFragmentInput
  path:  string   -- file path, relative to project root
  lines: string   -- line range, e.g. "150-210"

record ParsedRange
  start: integer  -- 1-indexed, inclusive
  end:   integer  -- 1-indexed, inclusive
```

---

## Functions

### ParseLineRange(lines) -> ParsedRange

Parses a line-range string such as `"150-210"` into a ParsedRange record.

```
function ParseLineRange(lines) -> ParsedRange
  errors:
    - "invalid line range": the format is wrong or the values are unusable.
```

1. Split `lines` on the `"-"` character.
   If the result does not have exactly two parts, raise error "invalid line range".

2. Parse the first part as integer `start`.
   Parse the second part as integer `end`.
   If either part is not a valid positive integer, raise error "invalid line range".

3. If `start` is less than 1, raise error "invalid line range".

4. If `start` is greater than `end`, raise error "invalid line range".

5. Return a ParsedRange with fields `start` and `end`.

---

### HashFragment(path, lines) -> string

Main entry point. Validates the path, reads the requested lines from the
file, and returns the SHA-1 base64url digest.

```
function HashFragment(path, lines) -> string
  errors:
    - "file not found":        the file does not exist at the given path.
    - "invalid line range":    the range format is invalid or out of bounds.
    - "path validation failure": the path is unsafe (traversal, absolute, etc.).
```

1. **Validate the path.**
   Call `ValidatePath(path, project_root)`.
   If it raises any error, re-raise it as "path validation failure: <original message>".

2. **Parse the line range.**
   Call `ParseLineRange(lines)` to obtain a ParsedRange record.
   If it raises "invalid line range", propagate the error unchanged.

3. **Open the file.**
   Call `OpenFileReader(path)` to obtain a FileReader.
   If the file cannot be opened (file unreadable / not found),
   raise error "file not found".

4. **Skip to the start line.**
   Call `SkipLines(reader, start - 1)` to advance past lines before
   the range.
   If "end of file" is raised during skipping, close the reader and
   raise error "invalid line range"
   (the file has fewer lines than `start - 1`).

5. **Read lines in the range.**
   Initialize an empty list `extracted`.
   For each index from `start` to `end` (inclusive):
     Call `ReadLine(reader)` to obtain the next line.
     If "end of file" is raised before the range is complete,
     close the reader and raise error "invalid line range"
     (the file has fewer lines than `end`).
     Append the line to `extracted`.

6. **Close the reader.**
   Call `Close(reader)`.

7. **Join the extracted lines.**
   Concatenate all lines in `extracted`, separated by `"\n"` (LF).
   The result is `content`.

8. **Compute the hash.**
   Compute the SHA-1 digest of `content` (treated as a byte sequence
   using its UTF-8 encoding).
   Encode the 20-byte digest as base64url (RFC 4648 §5) with no
   padding characters (`=`).
   The result is exactly 27 characters long.

9. **Return** the 27-character base64url string.

---

## Error summary

| Error                    | Condition                                                              |
|--------------------------|------------------------------------------------------------------------|
| `"path validation failure"` | Path is empty, absolute, contains `..`, or resolves outside root.  |
| `"invalid line range"`   | Format is wrong, start < 1, start > end, or end exceeds file length.  |
| `"file not found"`       | The file does not exist or cannot be opened.                           |

---

## Invariants

- Path validation always runs before the file is opened.
- Line endings in the extracted content are already normalized (LF only)
  because FileReader normalizes CRLF → LF before returning each line.
- The hash algorithm (SHA-1 + base64url, no padding, 27 characters) is
  identical to the algorithm used for the chain hash in `load_chain`,
  ensuring consistency across the tool framework.
- The tool is stateless: each call resolves its own inputs independently.
