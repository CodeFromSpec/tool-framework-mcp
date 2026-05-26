<!-- code-from-spec: ROOT/functional/utils/artifact_tag@HEcL7A1_AcX8jRpbXbw847bfxgg -->

# artifact_tag

Locates and extracts the `code-from-spec: <logical-name>@<hash>` tag from a
generated file. The tag may appear inside any comment syntax; the scanner does
not interpret comment delimiters — it searches every line for the substring.

---

## Data structures

```
record ArtifactTag
  logical_name: string   -- the part before the last "@"
  hash:         string   -- the 27-character base64url hash after the last "@"
```

---

## Functions

### ExtractArtifactTag(file_path) -> ArtifactTag

Open the file at `file_path` and scan it line-by-line until the artifact tag
is found. Close the file when done regardless of outcome.

**Parameters**
- `file_path` — path to the file to inspect.

**Returns**
- An `ArtifactTag` record with `logical_name` and `hash` populated.

**Errors**
- `"file unreadable"` — the file cannot be opened or read.
- `"no tag found"` — the file was fully read and no line contains the
  substring `code-from-spec: `.
- `"malformed tag"` — a line containing `code-from-spec: ` was found but the
  portion after the prefix does not satisfy the format rules (see step 5).

---

#### Step-by-step logic

1. Open the file at `file_path` using `OpenFileReader`.
   If the file cannot be opened, raise error `"file unreadable"`.

2. Set `tag_line` to empty.
   Set `found` to false.

3. Loop:
   a. Call `ReadLine` on the reader.
      If `ReadLine` raises `"end of file"`, exit the loop.
      If `ReadLine` raises any other read error, call `Close` on the reader,
      then raise error `"file unreadable"`.
   b. If the current line contains the substring `code-from-spec: `,
      set `tag_line` to the current line,
      set `found` to true,
      exit the loop.
      -- Stop at the first match; do not read the rest of the file.

4. Call `Close` on the reader.

5. If `found` is false, raise error `"no tag found"`.

6. Extract the raw tag value:
   a. Find the position of `code-from-spec: ` in `tag_line`.
   b. Take the substring starting immediately after `code-from-spec: `
      through the end of `tag_line`.
   c. Trim trailing whitespace from this substring.
      Call the result `raw`.

7. Locate the last occurrence of `@` in `raw`.
   If `@` is not present, raise error `"malformed tag"`.

8. Split `raw` at the last `@`:
   - `logical_name` = everything before the last `@`.
   - `hash`         = everything after the last `@`.

9. Validate:
   - If `logical_name` is empty, raise error `"malformed tag"`.
   - If `hash` is not exactly 27 characters long, raise error `"malformed tag"`.

10. Return an `ArtifactTag` record with
    - `logical_name` set to the value from step 8.
    - `hash`         set to the value from step 8.

---

## Contracts and invariants

- The file is read only until the first matching line; lines after the match
  are never read.
- The reader is always closed before the function returns (normal path or
  error path after the reader has been opened).
- A valid hash is always exactly 27 characters (base64url encoding).
- The split uses the **last** `@` so that a logical name containing `@`
  characters (none expected, but tolerated) does not break parsing.
