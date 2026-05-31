<!-- code-from-spec: ROOT/functional/logic/mcp_tools/hash_fragment@u-jaOuH4RrKUA8PbyyKuk8qbCD4 -->

# MCPHashFragment

function MCPHashFragment(path: string, lines: string) -> string
  errors:
    - InvalidLineRange: the range format is invalid,
      start < 1, start > end, or end exceeds the file's
      line count.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileReader.*): propagated from FileOpen.

  1. Call PathValidateCfs with path.
     If it raises an error, propagate it to the caller.

  2. Parse lines as a line range in the format "<start>-<end>".
     Both start and end are 1-based integers, inclusive.
     If the format does not match "<integer>-<integer>",
       raise error InvalidLineRange: "invalid line range format".
     If start < 1,
       raise error InvalidLineRange: "start line must be >= 1".
     If start > end,
       raise error InvalidLineRange: "start line must be <= end line".

  3. Create a PathCfs record with value set to path.
     Call FileOpen with that PathCfs.
     If FileOpen raises an error, propagate it to the caller.
     The result is a FileReader, call it reader.

  4. Call FileSkipLines(reader, start - 1) to skip past lines
     before the requested range.

  5. Set lines_to_read = end - start + 1.
     Create an empty list called collected_lines.
     For each index from 1 to lines_to_read:
       Call FileReadLine(reader).
       If FileReadLine raises EndOfFile:
         Call FileClose(reader).
         Raise error InvalidLineRange: "end line exceeds file line count".
       Append the returned line to collected_lines.

  6. Call FileClose(reader).

  7. Build the hash input string by taking each line in
     collected_lines in order and appending "\n" after it,
     including after the last line.

  8. Compute the SHA-1 digest of the hash input string (20 bytes).
     Encode the digest using base64url (RFC 4648 §5, no padding),
     producing a 27-character string.

  9. Return the 27-character base64url-encoded hash string.
