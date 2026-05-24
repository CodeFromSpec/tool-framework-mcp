---
outputs:
  - id: hash_fragment
    path: code-from-spec/functional/mcp_tools/hash_fragment/output.md
---

# ROOT/functional/mcp_tools/hash_fragment

Calculates the hash of a line range in a file, for use in
`external:` fragment declarations.

# Public

## Behavior

### Input

| Parameter | Required | Description |
|---|---|---|
| `path` | yes | File path relative to project root. |
| `lines` | yes | Line range (e.g., `"150-210"`). |

### Output

A text response containing the computed hash — a SHA-1 digest,
base64url encoded (RFC 4648 §5, no padding), 27 characters.

### Algorithm

1. Read the file at the given path.
2. Extract lines in the declared range (1-indexed, inclusive).
3. Normalize line endings: convert CRLF to LF.
4. Compute SHA-1 of the extracted content.
5. Encode as base64url without padding.

## Error conditions

| Condition | Description |
|---|---|
| File not found | The file does not exist. |
| Invalid line range | The range format is invalid or out of bounds. |
| Path validation failure | The path is unsafe (traversal, absolute, etc.). |
