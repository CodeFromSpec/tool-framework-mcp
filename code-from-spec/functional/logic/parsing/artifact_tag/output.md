<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@db-WIWK_SR9LGC-Xbjy5FNk5Gbk -->

record ArtifactTag
  logical_name: string
  hash: string


function ArtifactTagExtract(file_path: pathutils.PathCfs) -> ArtifactTag
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - NoTagFound: the file has no "code-from-spec:" substring.
    - MalformedTag: the tag exists but cannot be parsed
      (no @, empty name, wrong hash length).
    - (FileReader.*): propagated from FileOpen.

  1. Call FileOpen(file_path) to obtain a reader.
     If FileOpen raises FileUnreadable or any PathUtils error,
     propagate it to the caller.

  2. Set found_line to empty.
     Set done to false.

  3. Loop until done is true:
     a. Call FileReadLine(reader) to get the next line.
        If FileReadLine raises EndOfFile, set done to true and exit loop.
     b. If the line contains the substring "code-from-spec: ",
        set found_line to that line and set done to true.

  4. Call FileClose(reader).

  5. If found_line is empty, raise error "NoTagFound".

  6. Find the position of "code-from-spec: " within found_line.
     Take the substring starting immediately after "code-from-spec: ".
     Trim leading whitespace from that substring. Call it raw_tag.

  7. Find the first occurrence of "@" in raw_tag.
     If "@" is not found, raise error "MalformedTag".

  8. Set logical_name to the portion of raw_tag before "@".
     If logical_name is empty, raise error "MalformedTag".

  9. Set hash_candidate to the portion of raw_tag immediately after "@".
     If hash_candidate has fewer than 27 characters, raise error "MalformedTag".
     Set hash to the first 27 characters of hash_candidate.

  10. Return ArtifactTag with:
        logical_name = logical_name
        hash = hash
