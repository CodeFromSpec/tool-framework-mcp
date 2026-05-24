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

## Behavior

### Input

A byte sequence (text content).

### Output

A 27-character string: SHA-1 digest encoded as base64url
(RFC 4648 §5, no padding).

## Normalization

Before hashing, normalize the content:
- Convert CRLF line endings to LF.
- No other normalization.

## Line range hashing

Given a file path and a line range (e.g., `"150-210"`):
1. Read the file.
2. Extract lines in the range (1-indexed, inclusive).
3. Normalize and hash the extracted content.

This is used for `external:` fragment hash verification
and by the `hash_fragment` tool.

## Contracts

- Deterministic — same content always produces same hash.
- The 27-character output is the canonical text
  representation used in artifact tags, fragment hashes,
  and chain hashes throughout the framework.
