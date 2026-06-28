---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/artifact_tag
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
output: internal/artifacttag/artifacttag.go
---

# SPEC/golang/implementation/parsing/artifact_tag

# Agent

Implement the artifact tag extraction as a Go package.

## Logic

1. Call `FileOpen(file_path, "read", 30000)`.
   If `FileOpen` raises an error, propagate it.
   Store the result as `handle`.

2. Set `tag_line` to empty (not yet found).

3. Loop:
   a. Call `FileReadLine(handle)`.
      If it raises `EndOfFile`, exit the loop.
      If it raises any other error, call
      `FileClose(handle)` then propagate the error.
   b. Store the returned line as `line`.
   c. If `line` contains the substring
      `"code-from-spec: "`: set `tag_line` to `line`
      and exit the loop.

4. Call `FileClose(handle)`.

5. If `tag_line` is empty: raise error `NoTagFound`.

6. Find the index of `"code-from-spec: "` within
   `tag_line`. Take the substring starting immediately
   after that occurrence. Store it as `remainder`.

7. Trim leading whitespace from `remainder`.

8. Find the index of the first `"@"` in `remainder`.
   If `"@"` is not found: raise error `MalformedTag`.

9. Set `logical_name` to the substring of `remainder`
   from position 0 up to (not including) the `"@"`.
   If `logical_name` is empty: raise error
   `MalformedTag`.

10. Set `after_at` to the substring of `remainder`
    starting immediately after `"@"`.
    If the length of `after_at` is less than 27:
    raise error `MalformedTag`.

11. Set `hash` to the first 27 characters of `after_at`.

12. Return `ArtifactTag` with:
    - `logical_name` = `logical_name`
    - `hash` = `hash`

## Go-specific guidance

- Use the `file` package to open and read the file
  line by line.
- Error sentinels with `errors.New`.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
- Scan for the `code-from-spec: ` substring in each line.
  Stop reading as soon as a match is found.
- Parse the tag by finding the first `@` after the prefix.
- The hash is exactly the first 27 characters after `@`.
  Anything after those 27 characters is ignored.
