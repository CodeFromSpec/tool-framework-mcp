<!-- code-from-spec: ROOT/functional/logic/mcp_tools/hash_fragment@Uq_bE8-zugR_gWNeqkQUKTCEs48 -->

# MCPHashFragment

function MCPHashFragment(path: string, lines: string) -> string
  errors:
    - InvalidLineRange: the range format is invalid,
      start < 1, start > end, or end exceeds the file's
      line count.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileReader.*): propagated from FileOpen.

  1. **Validate path**
     Call PathValidateCfs with path.
     If it fails, propagate the error.

  2. **Parse line range**
     Parse lines as the pattern <start>-<end> where both values are integers.
     If the pattern does not match, raise error InvalidLineRange.
     Convert start and end to integers.
     If start < 1, raise error InvalidLineRange.
     If start > end, raise error InvalidLineRange.

  3. **Read lines**
     Create a PathCfs record with value set to path.
     Call FileOpen with that PathCfs.
     If FileOpen fails, propagate the error.
     Call FileSkipLines with the reader and count = start - 1.
     Set lines_to_read = end - start + 1.
     Create an empty list called collected_lines.
     Repeat lines_to_read times:
       Call FileReadLine with the reader.
       If FileReadLine raises EndOfFile:
         Call FileClose with the reader.
         Raise error InvalidLineRange.
       Append the returned line to collected_lines.
     Call FileClose with the reader.

  4. **Compute hash**
     Build a single text block by taking each line in collected_lines
     and appending "\n" after it (including after the last line).
     Compute the SHA-1 digest of the resulting text block (as bytes,
     using UTF-8 encoding).
     Encode the 20-byte digest using base64url encoding
     (RFC 4648 §5, no padding), producing a 27-character string.
     Return the 27-character hash string.
