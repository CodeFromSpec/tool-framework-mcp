<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@OgAUE8S-b0AXBnNjvW2stFWNMhk -->

## Records

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

```
function ArtifactTagExtract(file_path: pathutils.PathCfs) -> ArtifactTag
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - NoTagFound: the file has no "code-from-spec: " substring.
    - MalformedTag: the tag exists but cannot be parsed
      (no @, empty logical name, or fewer than 27 characters after @).
    - (FileReader.*): propagated from FileOpen.
```

### ArtifactTagExtract

  1. Call `FileOpen` with `file_path`.
     If `FileOpen` raises `FileUnreadable` or a `PathUtils.*` error, propagate it.

  2. Read lines one at a time using `FileReadLine`.
     For each line:
     a. Check if the line contains the substring `"code-from-spec: "`.
        If not, continue to the next line.
     b. If the substring is found, call `FileClose` and proceed to step 3.
     If `FileReadLine` raises `EndOfFile` before a match is found:
       Call `FileClose`.
       Raise error `NoTagFound`.

  3. Take the portion of the line starting immediately after `"code-from-spec: "`.
     Trim leading whitespace from this portion.

  4. Find the first occurrence of `"@"` in the trimmed portion.
     If `"@"` is not found, raise error `MalformedTag`.

  5. Extract the logical name: everything from the start of the trimmed portion
     up to (but not including) `"@"`.
     If the logical name is empty, raise error `MalformedTag`.

  6. Extract the hash: the 27 characters immediately after `"@"`.
     If fewer than 27 characters remain after `"@"`, raise error `MalformedTag`.

  7. Return an `ArtifactTag` record with:
     - `logical_name`: the extracted logical name
     - `hash`: the 27-character hash
