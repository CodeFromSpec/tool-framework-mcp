<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@idM2UINxtdjvCeKsBBTQSRbb1hg -->

## Data Structures

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

---

### ArtifactTagExtract(file_path: PathCfs) -> ArtifactTag

Opens the file at `file_path`, scans line by line for the first occurrence
of the artifact tag pattern, extracts the logical name and hash, then
closes the file.

**Errors:**
- `(path errors)`: propagated from FileOpen.
- `file unreadable`: the file cannot be opened or read.
- `no tag found`: no line in the file contains the substring `"code-from-spec: "`.
- `malformed tag`: a matching line was found but the tag cannot be parsed
  (missing `@`, empty logical name, or fewer than 27 characters after `@`).

**Steps:**

1. Call `FileOpen(file_path)` to obtain a `FileReader`.
   If `FileOpen` raises a path error, propagate it.
   If `FileOpen` raises a file-unreadable error, raise error `"file unreadable"`.

2. Set `found_line` to empty (no value yet).

3. Loop:
   a. Call `FileReadLine(reader)`.
      If it raises `"end of file"`, exit the loop.
   b. If the current line contains the substring `"code-from-spec: "`:
      - Set `found_line` to the current line.
      - Exit the loop immediately (do not read further lines).

4. Call `FileClose(reader)`.

5. If `found_line` has no value:
   Raise error `"no tag found"`.

6. Extract the tag value from `found_line`:
   a. Find the position of the substring `"code-from-spec: "` in `found_line`.
   b. Take the substring of `found_line` starting immediately after
      `"code-from-spec: "`.
   c. Trim leading whitespace from this substring. Call it `tag_value`.

7. Find the first occurrence of `"@"` in `tag_value`.
   If `"@"` is not present:
   Raise error `"malformed tag"`.

8. Set `logical_name` to everything in `tag_value` before the first `"@"`.
   If `logical_name` is empty:
   Raise error `"malformed tag"`.

9. Set `after_at` to everything in `tag_value` after the first `"@"`.
   If `after_at` has fewer than 27 characters:
   Raise error `"malformed tag"`.

10. Set `hash` to the first 27 characters of `after_at`.

11. Return an `ArtifactTag` record with:
    - `logical_name`: the value from step 8.
    - `hash`: the value from step 10.
```

## Contracts

- The file is read only until the first matching line. Lines after the match
  are never read.
- The hash field of the returned `ArtifactTag` is always exactly 27 characters.
- `FileClose` is always called, regardless of whether a tag was found or an
  error occurred during extraction.
