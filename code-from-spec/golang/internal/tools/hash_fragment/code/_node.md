---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/internal/file_reader
  - ROOT/golang/internal/pathvalidation
outputs:
  - id: hash_fragment
    path: internal/hash_fragment/hash_fragment.go
---

# ROOT/golang/internal/tools/hash_fragment/code

Implementation of the hash_fragment tool handler.

# Agent

## Implementation

1. Validate `args.Path` using `pathvalidation.ValidatePath`
   against the working directory. If it fails, return a tool
   error with the validation error.
2. Parse `args.Lines` as a `"start-end"` range where start
   and end are 1-indexed, inclusive integers. If the format
   is invalid (not two integers separated by a hyphen, or
   start > end, or start < 1), return a tool error:
   `"invalid line range: <lines>"`.
3. Read the file using `filereader`. If the file does not
   exist or cannot be read, return a tool error.
4. Extract lines from `start` to `end` (1-indexed, inclusive).
   If `end` exceeds the file's line count, return a tool
   error: `"invalid line range: <lines> (file has <n> lines)"`.
5. Join the extracted lines with LF (`\n`).
6. Compute SHA-1 of the joined content using `crypto/sha1`.
7. Encode the hash as base64url (RFC 4648 S5, no padding)
   using `encoding/base64.RawURLEncoding` -- produces a
   27-character string.
8. Return the hash as a success result text.

### Error handling

- Path validation failure -> tool error with the violation.
- Invalid line range format -> tool error:
  `"invalid line range: <lines>"`.
- File not found -> tool error: `"file not found: <path>"`.
- Line range out of bounds -> tool error:
  `"invalid line range: <lines> (file has <n> lines)"`.

## Constraints

- The path must pass path validation before reading.
- Line endings are already normalized by filereader.
- Uses `crypto/sha1` and `encoding/base64` (RawURLEncoding,
  no padding), same algorithm as the chain hash in
  load_chain.
