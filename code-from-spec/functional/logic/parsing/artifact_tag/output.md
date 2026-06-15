<!-- code-from-spec: SPEC/functional/logic/parsing/artifact_tag@Thn4Qw2HqQ6_ck8SOaig6aU40mM -->

namespace: artifacttag

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
     If FileOpen raises any error, propagate it.

  2. Set found_line to empty.
     Set reader_open to true.

  3. Loop:
     a. Call FileReadLine(reader).
        If EndOfFile is raised, break out of the loop.
     b. If the line contains the substring "code-from-spec: ":
        Set found_line to this line.
        Break out of the loop.

  4. Call FileClose(reader).

  5. If found_line is empty, raise error NoTagFound.

  6. Find the position of "code-from-spec: " in found_line.
     Take the substring starting immediately after "code-from-spec: ".
     Trim leading whitespace from this substring.
     Call this remainder.

  7. Find the first occurrence of "@" in remainder.
     If "@" is not found, raise error MalformedTag.

  8. Set logical_name to the substring of remainder from the start
     up to (but not including) the "@".
     If logical_name is empty, raise error MalformedTag.

  9. Set hash_candidate to the substring of remainder starting
     immediately after the "@".
     If hash_candidate has fewer than 27 characters, raise error MalformedTag.
     Set hash to the first 27 characters of hash_candidate.

  10. Return ArtifactTag with:
        logical_name = logical_name
        hash = hash
