---
depends_on:
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/os/path_utils
outputs:
  - id: hash_fragment
    path: code-from-spec/functional/logic/mcp_tools/hash_fragment/output.md
---

# ROOT/functional/logic/mcp_tools/hash_fragment

Calculates the hash of a line range in a file, for use
in `external:` fragment declarations.

# Public

## Interface

```
function MCPHashFragment(path: string, lines: string) -> string
  errors:
    - InvalidLineRange: the range format is invalid,
      start < 1, start > end, or end exceeds the file's
      line count.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileReader.*): propagated from FileOpen.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `path` | yes | File path relative to project root (forward slashes). |
| `lines` | yes | Line range (e.g., `"150-210"`). |

### Output

A text response containing the computed hash — a SHA-1
digest, base64url encoded (RFC 4648 §5, no padding),
27 characters.

# Agent

## Behavior

### Step 1 — Validate path

Call `PathValidateCfs` with the `path` string. If it
fails, propagate the error.

### Step 2 — Parse line range

Parse `lines` as `start-end` (both 1-based, inclusive).
Example: `"150-210"` means lines 150 through 210. If
the format is invalid, `start < 1`, or `start > end`,
raise InvalidLineRange.

### Step 3 — Read lines

Create a `PathCfs` from the `path` string. Open the
file with `FileOpen`. If it fails, propagate the error.

Use `FileSkipLines` to skip `start - 1` lines, then
read `end - start + 1` lines with `FileReadLine`. If
`FileReadLine` returns EndOfFile before all lines are
read, call `FileClose` and raise InvalidLineRange
(end exceeds the file's line count).

Call `FileClose`.

### Step 4 — Compute hash

Append `\n` after each line, including the last (per
the framework hashing convention). Compute SHA-1 of
the result. Encode the 20-byte digest as base64url
(RFC 4648 §5, no padding) — 27 characters.

Return the hash string.

## Contracts

- The path must pass `PathValidateCfs` before reading.
- Line endings are already normalized by `FileReadLine`.
- Uses SHA-1 + base64url (no padding), same algorithm
  and hashing convention as `ChainHashCompute`.
