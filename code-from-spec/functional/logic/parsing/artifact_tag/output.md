<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@aOIPH93sQqs_EkIYXBx5pMcmYVk -->

# ArtifactTag

## Records

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

---

### ArtifactTagExtract(file_path) -> ArtifactTag

Parameters:
- file_path: PathCfs — path to the file to scan

Returns: ArtifactTag record with fields logical_name and hash

Errors:
- (path errors): propagated from FileOpen
- "file unreadable": the file cannot be opened or read
- "no tag found": the file contains no "code-from-spec: " substring
- "malformed tag": the tag exists but cannot be parsed (no @ character, empty logical name, or fewer than 27 characters after @)

Steps:

  1. Call FileOpen with file_path.
     If FileOpen raises a path error, propagate the error.
     If the file cannot be opened, raise error "file unreadable".

  2. Set found_line to empty (no match yet).

  3. Loop:
     a. Call FileReadLine on the reader.
        If FileReadLine raises "end of file", exit the loop.
     b. If the current line contains the substring "code-from-spec: ":
        Set found_line to the current line.
        Exit the loop.

  4. Call FileClose on the reader.

  5. If found_line is empty (no match was found):
     Raise error "no tag found".

  6. Take the portion of found_line starting immediately after
     the first occurrence of "code-from-spec: ".
     Call this raw_tag.

  7. Trim leading whitespace from raw_tag.

  8. Find the first occurrence of "@" in raw_tag.
     If "@" is not found:
       Raise error "malformed tag".

  9. Extract the logical name as the substring from the start of
     the trimmed raw_tag up to (but not including) the "@".
     If the logical name is empty:
       Raise error "malformed tag".

  10. Extract the hash as the 27 characters immediately after the "@".
      If there are fewer than 27 characters after "@":
        Raise error "malformed tag".

  11. Return an ArtifactTag record with:
        logical_name = the extracted logical name
        hash         = the extracted 27-character hash
```
