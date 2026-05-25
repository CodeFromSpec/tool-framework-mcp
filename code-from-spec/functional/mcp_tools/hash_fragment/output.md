<!-- code-from-spec: ROOT/functional/mcp_tools/hash_fragment@PENDING -->

## Functions

### function HashFragment(path, lines) -> string

Calculates the SHA-1 hash of a line range in a file, for use
in external fragment declarations.

**Parameters**

- path: string, required. File path relative to project root.
- lines: string, required. Line range in "start-end" format
  (1-indexed, inclusive).

**Algorithm**

1. Validate the path using path_validation.
   If the path is unsafe (empty, absolute, contains traversal,
   or escapes the project root), raise error "path validation failure".

2. Open the file at path using file_reader.
   If the file does not exist, raise error "file not found".

3. Parse the lines parameter as "start-end" where start and end
   are integers.
   If the format is invalid, raise error "invalid line range".

4. If start is greater than end, raise error "invalid line range".

5. If end exceeds the file's total line count,
   raise error "invalid line range".

6. Extract lines from start to end (1-indexed, inclusive).

7. Join the extracted lines with LF (linefeed character).

8. Compute the SHA-1 digest of the joined content.

9. Encode the SHA-1 digest as base64url (RFC 4648 section 5,
   no padding), producing a 27-character string.

10. Return the 27-character hash string.
