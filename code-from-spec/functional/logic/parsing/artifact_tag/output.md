<!-- code-from-spec: SPEC/functional/logic/parsing/artifact_tag@5Px8Sc3W12KCZUoQFgHCML80fmE -->

# artifact_tag

## Records

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

### ArtifactTagExtract

```
function ArtifactTagExtract(file_path: pathutils.PathCfs) -> ArtifactTag
```

Parameters:
- `file_path` — CFS path to the file to scan

Returns: `ArtifactTag` with `logical_name` and `hash` fields

Errors:
- `NoTagFound`: no line in the file contains the substring `"code-from-spec: "`
- `MalformedTag`: a matching line was found but the tag cannot be parsed (`@` is absent, logical name is empty, or fewer than 27 characters follow `@`)
- `(File.*)`: propagated from `FileOpen` or `FileReadLine`

Steps:

1. Call `FileOpen(file_path, "read", 30000)`.
   If `FileOpen` raises an error, propagate it.
   Store the result as `handle`.

2. Set `tag_line` to empty (not yet found).

3. Loop:
   a. Call `FileReadLine(handle)`.
      If it raises `EndOfFile`, exit the loop.
      If it raises any other error, call `FileClose(handle)` then propagate the error.
   b. Store the returned line as `line`.
   c. If `line` contains the substring `"code-from-spec: "`:
      set `tag_line` to `line` and exit the loop.

4. Call `FileClose(handle)`.

5. If `tag_line` is empty:
   raise error `NoTagFound`.

6. Find the index of `"code-from-spec: "` within `tag_line`.
   Take the substring of `tag_line` starting immediately after that occurrence.
   Store it as `remainder`.

7. Trim leading whitespace from `remainder`.

8. Find the index of the first `"@"` in `remainder`.
   If `"@"` is not found:
     raise error `MalformedTag`.

9. Set `logical_name` to the substring of `remainder` from position 0 up to (not including) the `"@"`.
   If `logical_name` is empty:
     raise error `MalformedTag`.

10. Set `after_at` to the substring of `remainder` starting immediately after `"@"`.
    If the length of `after_at` is less than 27:
      raise error `MalformedTag`.

11. Set `hash` to the first 27 characters of `after_at`.

12. Return `ArtifactTag` with:
    - `logical_name` = `logical_name`
    - `hash` = `hash`
