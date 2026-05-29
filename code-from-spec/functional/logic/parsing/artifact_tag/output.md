<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@idM2UINxtdjvCeKsBBTQSRbb1hg -->

# artifact_tag

## Records

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

---

### ArtifactTagExtract(file_path: PathCfs) -> ArtifactTag

Opens the file at `file_path` and scans line by line for an artifact tag.
Stops at the first matching line. Always closes the reader before returning.

**Errors:**
- `(path errors)`: propagated from FileOpen.
- `"file unreadable"`: the file cannot be opened or read.
- `"no tag found"`: the file has no `code-from-spec: ` substring.
- `"malformed tag"`: the tag exists but cannot be parsed (no `@`, empty name, or wrong hash length).

**Steps:**

1. Call `FileOpen(file_path)` to obtain a reader.
   If FileOpen raises an error, propagate it to the caller.

2. Set `found_line` to empty.
   Set `read_error` to none.

3. Loop:
   a. Call `FileReadLine(reader)`.
      If it raises "end of file", exit the loop.
      If it raises any other error, set `read_error` to that error, exit the loop.
   b. If the line contains the substring `"code-from-spec: "`, set `found_line` to that line
      and exit the loop.

4. Call `FileClose(reader)`.

5. If `read_error` is set, raise error `"file unreadable"`.

6. If `found_line` is empty, raise error `"no tag found"`.

7. Extract the tag value from `found_line`:
   a. Find the index of `"code-from-spec: "` in the line.
   b. Take the substring starting immediately after `"code-from-spec: "`.
   c. Trim leading whitespace from that substring. Call this `raw_tag`.

8. Find the first occurrence of `"@"` in `raw_tag`.
   If `"@"` is not found, raise error `"malformed tag"`.

9. Set `logical_name` to the substring of `raw_tag` from the start up to (but not including) `"@"`.
   If `logical_name` is empty, raise error `"malformed tag"`.

10. Set `remainder` to the substring of `raw_tag` immediately after `"@"`.
    If `remainder` has fewer than 27 characters, raise error `"malformed tag"`.

11. Set `hash` to the first 27 characters of `remainder`.

12. Return an `ArtifactTag` record with:
    - `logical_name`: the value from step 9.
    - `hash`: the value from step 11.
```
