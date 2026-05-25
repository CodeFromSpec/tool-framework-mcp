---
depends_on:
  - ROOT/functional/utils/file_reader
  - ROOT/functional/utils/path_validation
outputs:
  - id: hash_fragment
    path: code-from-spec/functional/mcp_tools/hash_fragment/output.md
---

# ROOT/functional/mcp_tools/hash_fragment

Calculates the hash of a line range in a file, for use in
`external:` fragment declarations.

# Public

## Interface

```
function HashFragment(path, lines) -> string
  errors:
    - file not found: the file does not exist.
    - invalid line range: the range format is invalid or out of bounds.
    - path validation failure: the path is unsafe (traversal, absolute, etc.).
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `path` | yes | File path relative to project root. |
| `lines` | yes | Line range (e.g., `"150-210"`). |

### Output

A text response containing the computed hash — a SHA-1 digest,
base64url encoded (RFC 4648 S5, no padding), 27 characters.

# Agent

## Behavior

### Algorithm

1. Validate the path using path validation.
2. Open the file with file_reader.
3. Skip to the start of the line range.
4. Read lines in the range (1-indexed, inclusive).
5. Join the extracted lines with LF.
6. Compute SHA-1 of the joined content.
7. Encode as base64url (RFC 4648 §5, no padding) — 27 characters.

### Line range format

`"start-end"` where start and end are 1-indexed, inclusive.
If start > end, or end exceeds the file's line count, raise
"invalid line range".

## Contracts

- The path must pass path validation before reading.
- Line endings are already normalized by file_reader.
- Uses SHA-1 + base64url (no padding), same algorithm as
  the chain hash in load_chain.
