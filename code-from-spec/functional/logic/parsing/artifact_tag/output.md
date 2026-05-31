<!-- code-from-spec: ROOT/functional/logic/parsing/artifact_tag@npVvrSV44M4zMIVGc6HwdynaRTA -->

## ArtifactTag

A record representing a parsed artifact tag found in a generated file.

Fields:
- logical_name: string — the logical name portion of the tag
- hash: string — the 27-character base64url hash portion of the tag


## ArtifactTagExtract(file_path) -> ArtifactTag

Parameters:
- file_path: PathCfs — path to the file to scan

Returns: ArtifactTag

Errors:
- FileUnreadable: the file cannot be opened or read
- NoTagFound: the file contains no "code-from-spec: " substring
- MalformedTag: a tag was found but cannot be parsed (missing "@", empty logical name, or fewer than 27 characters after "@")
- (FileReader.*): propagated from FileOpen

Steps:

1. Open the file at file_path using FileOpen.
   If FileOpen raises an error, propagate it.
   Bind the result to reader.

2. Set found_line to empty (no match yet).

3. Loop:
   a. Call FileReadLine(reader).
      If EndOfFile is raised, exit the loop.
   b. If the current line contains the substring "code-from-spec: ":
      Set found_line to this line.
      Exit the loop.

4. Call FileClose(reader).

5. If found_line is empty (no match was found):
   Raise error NoTagFound.

6. Take the substring of found_line starting immediately after
   the first occurrence of "code-from-spec: ".
   Trim leading whitespace from this substring.
   Bind the result to tag_content.

7. Find the index of the first occurrence of "@" in tag_content.
   If "@" is not found:
     Raise error MalformedTag "tag missing '@' separator".

8. Extract logical_name as the substring of tag_content from
   position 0 up to (but not including) the "@".
   If logical_name is empty:
     Raise error MalformedTag "logical name is empty".

9. Extract hash_candidate as the substring of tag_content
   starting immediately after "@".
   If the length of hash_candidate is less than 27:
     Raise error MalformedTag "hash must be at least 27 characters".

10. Set hash to the first 27 characters of hash_candidate.

11. Return ArtifactTag with:
    - logical_name: logical_name
    - hash: hash
