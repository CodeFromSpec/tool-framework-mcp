<!-- code-from-spec: ROOT/functional/utils/artifact_tag@PENDING -->

## Data structures

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

### ExtractArtifactTag(file_path) -> ArtifactTag

1. Open the file at file_path using file_reader.
   If the file cannot be opened or read,
   raise error "file unreadable".

2. Read the file line by line using ReadLine.

3. For each line:
   a. Check whether the line contains the substring "code-from-spec: ".
   b. If not found, continue to the next line.
   c. If found, proceed to step 4.

4. If end of file is reached without finding the substring,
   raise error "no tag found".

5. Take everything after the first occurrence of "code-from-spec: "
   to the end of the line. Trim trailing whitespace. Call this
   the tag_value.

6. Find the last occurrence of "@" in tag_value.
   If "@" is not found, raise error "malformed tag".

7. Set logical_name to everything before the last "@".
   Set hash to everything after the last "@".

8. If logical_name is empty, raise error "malformed tag".

9. If the length of hash is not exactly 27 characters,
   raise error "malformed tag".

10. Return an ArtifactTag record with logical_name and hash.
