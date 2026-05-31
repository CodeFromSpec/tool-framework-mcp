<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@XY0chhvi-N9_etVVcJpmVaxRFo8 -->

# artifact_tag

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
      (no @, empty name, wrong hash length).
    - (FileReader.*): propagated from FileOpen.
```

### ArtifactTagExtract

  1. Call FileOpen(file_path) to open the file for reading.
     If FileOpen raises FileUnreadable or any PathUtils error,
     propagate the error to the caller without further processing.

  2. Set found_line to empty.
     Set reader_done to false.

  3. Loop:
     Call FileReadLine(reader) to get the next line.
     If FileReadLine raises EndOfFile, set reader_done to true and break the loop.
     If the line contains the substring "code-from-spec: ",
       set found_line to this line and break the loop.

  4. Call FileClose(reader).

  5. If found_line is empty,
     raise error "NoTagFound: the file has no code-from-spec: tag".

  6. Find the position of "code-from-spec: " in found_line.
     Take the substring starting immediately after "code-from-spec: ".
     Call this raw_tag.

  7. Trim leading whitespace from raw_tag.

  8. Find the first occurrence of "@" in raw_tag.
     If "@" is not found,
       raise error "MalformedTag: no @ separator found in tag".

  9. Extract logical_name as everything before the first "@" in raw_tag.
     If logical_name is empty,
       raise error "MalformedTag: logical name is empty".

  10. Extract the substring immediately after "@".
      Call this hash_candidate.
      If the length of hash_candidate is less than 27,
        raise error "MalformedTag: hash must be at least 27 characters".

  11. Set hash to the first 27 characters of hash_candidate.

  12. Return ArtifactTag with
        logical_name set to the extracted logical_name
        hash set to hash.
```
