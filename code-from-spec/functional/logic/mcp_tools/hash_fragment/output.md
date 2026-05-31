<!-- code-from-spec: ROOT/functional/logic/mcp_tools/hash_fragment@Cv4svwS9lP9fKBeuL5pcUWEK5FE -->

# MCPHashFragment

function MCPHashFragment(path: string, lines: string) -> string
  errors:
    - InvalidLineRange: the range format is invalid,
      start < 1, start > end, or end exceeds the file's
      line count.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileReader.*): propagated from FileOpen.

  1. Validate path.
     Call PathValidateCfs with path.
     If it raises an error, propagate it to the caller.

  2. Parse line range.
     Split lines on the character "-".
     If the result does not have exactly two parts, raise InvalidLineRange
       with message "invalid line range format: expected <start>-<end>".
     Parse the first part as integer start.
     Parse the second part as integer end.
     If either part is not a valid integer, raise InvalidLineRange
       with message "invalid line range format: start and end must be integers".
     If start < 1, raise InvalidLineRange
       with message "invalid line range: start must be >= 1".
     If start > end, raise InvalidLineRange
       with message "invalid line range: start must be <= end".

  3. Read lines from the file.
     Create a PathCfs record with value set to path.
     Call FileOpen with that PathCfs.
     If FileOpen raises an error, propagate it to the caller.
     The result is reader.

     Call FileSkipLines with reader and count = start - 1.

     Set line_count = end - start + 1.
     Initialize collected_lines as an empty list of strings.
     For each index from 1 to line_count:
       Call FileReadLine with reader.
       If FileReadLine raises EndOfFile:
         Call FileClose with reader.
         Raise InvalidLineRange with message
           "invalid line range: end exceeds the file's line count".
       Append the returned line to collected_lines.

     Call FileClose with reader.

  4. Compute the hash.
     Initialize content as an empty string.
     For each line in collected_lines:
       Append line to content.
       Append "\n" to content.
     Compute the SHA-1 digest of content (treated as a byte sequence in UTF-8).
     Encode the 20-byte digest as base64url (RFC 4648 §5, no padding).
     The result is a 27-character string.

     Return the hash string.
