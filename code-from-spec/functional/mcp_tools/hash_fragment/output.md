<!-- code-from-spec: ROOT/functional/mcp_tools/hash_fragment@3FMTkXZ8aKl1dmsIawjKGQsEV78 -->

# HashFragment

Computes a SHA-1 / base64url hash over a contiguous slice of lines
extracted from a file.  The result is a 27-character base64url string
(RFC 4648 §5, no padding).

---

## Dependencies

Uses the following shared interfaces:

- **ValidatePath(relative_path, project_root)** — rejects empty, absolute,
  traversal, or out-of-root paths.
- **OpenFileReader(file_path)** — opens a file for sequential line reading.
- **ReadLine(reader)** — returns the next line (CRLF already normalized to LF),
  raises "end of file" when exhausted.
- **SkipLines(reader, count)** — discards `count` lines without returning them.

---

## ParseLineRange(lines) -> (start, end)

Parses the `lines` parameter string into a (start, end) integer pair.

```
1. Split lines on the "-" character.
   If the result does not have exactly 2 parts, raise error "invalid line range".

2. Parse the first part as an integer -> start.
   Parse the second part as an integer -> end.
   If either part is not a valid positive integer, raise error "invalid line range".

3. If start < 1 or end < 1, raise error "invalid line range".
   If start > end, raise error "invalid line range".

4. Return (start, end).
```

---

## HashFragment(path, lines) -> string

Main entry point.  Validates the path, extracts the requested lines,
and returns the SHA-1 / base64url digest.

```
1. Call ValidatePath(path, project_root).
   If it raises any error, propagate that error as "path validation failure:
   <original error message>".

2. Call ParseLineRange(lines) -> (start, end).
   If it raises "invalid line range", propagate the error unchanged.

3. Call OpenFileReader(path).
   If the file cannot be opened, raise error "file not found".
   Let reader be the returned FileReader.

4. Skip to the start line.
   Call SkipLines(reader, start - 1).
   If "end of file" is raised before skipping is complete, raise error
   "invalid line range".

5. Read lines from start to end (inclusive).
   Initialize collected_lines as an empty list.
   Repeat (end - start + 1) times:
     Call ReadLine(reader) -> line.
     If "end of file" is raised, raise error "invalid line range".
     Append line to collected_lines.

6. Join collected_lines with LF ("\n") -> content.
   (Do NOT append a trailing LF after the last line.)

7. Compute SHA-1 of content (treating the joined string as UTF-8 bytes).

8. Encode the 20-byte SHA-1 digest as base64url (RFC 4648 §5, no padding).
   The result is exactly 27 characters.

9. Return the 27-character base64url string.
```

---

## Error conditions

| Error | Trigger |
|---|---|
| `"file not found"` | The file at `path` does not exist or cannot be opened. |
| `"invalid line range"` | The `lines` string is malformed, start > end, or end exceeds the file's line count. |
| `"path validation failure: <reason>"` | ValidatePath rejected the path (empty, absolute, traversal, outside root). |

---

## Invariants

- Line numbers are 1-indexed and inclusive on both ends.
- Line endings are normalized to LF by file_reader before hashing.
- The hash algorithm (SHA-1 + base64url no-padding) is identical to the
  one used for chain hashes in load_chain, so results are directly comparable.
- The output string is always exactly 27 characters.
```
