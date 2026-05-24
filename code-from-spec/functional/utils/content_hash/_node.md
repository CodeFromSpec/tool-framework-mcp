---
outputs:
  - id: content_hash
    path: code-from-spec/functional/utils/content_hash/output.md
---

# ROOT/functional/utils/content_hash

Computes SHA-1 hashes of content, represented as base64url
strings. This is the fundamental hashing primitive used
throughout the framework.

# Public

## Interface

```
function ContentHash(content) -> string

function HashLineRange(file_path, line_range) -> string
  errors:
    - file unreadable: the file cannot be opened or read.
    - invalid line range: the range is out of bounds or malformed.
```

ContentHash takes a byte sequence and returns a 27-character
base64url string (SHA-1 digest, RFC 4648 S5, no padding).

HashLineRange reads a file, extracts lines in the given range
(1-indexed, inclusive), normalizes, and hashes the content.
Used for `external:` fragment hash verification and by the
`hash_fragment` tool.

# Agent

## Behavior

### Normalization

Before hashing, normalize the content:
- Convert CRLF line endings to LF.
- No other normalization.

### Line range hashing

Given a file path and a line range (e.g., `"150-210"`):
1. Read the file.
2. Extract lines in the range (1-indexed, inclusive).
3. Normalize and hash the extracted content.

## Contracts

- Deterministic — same content always produces same hash.
- The 27-character output is the canonical text
  representation used in artifact tags, fragment hashes,
  and chain hashes throughout the framework.
