---
outputs:
  - id: hash_fragment_test
    path: internal/hash_fragment/hash_fragment_test.go
---

# ROOT/golang/internal/tools/hash_fragment/tests

Tests for the hash_fragment tool handler.

# Agent

## Context

Each test uses `t.TempDir()` as the project root and working
directory. A test file with known content is created so that
expected hashes can be precomputed.

## Happy Path

### Hashes a valid line range

Create a file with multiple lines of known content. Call the
handler with `Path` pointing to the file and `Lines: "2-4"`.

Expect: success result. The text contains a 27-character
base64url-encoded SHA-1 hash. Verify the hash matches the
expected SHA-1 of the joined lines 2 through 4 (inclusive,
joined with LF).

### Single line range

Create a file with multiple lines. Call the handler with
`Lines: "3-3"`.

Expect: success result. The hash matches the SHA-1 of just
line 3 (no trailing LF from joining).

### First line of file

Create a file with multiple lines. Call the handler with
`Lines: "1-1"`.

Expect: success result. The hash matches the SHA-1 of the
first line only.

### Last line of file

Create a file with exactly 5 lines. Call the handler with
`Lines: "5-5"`.

Expect: success result. The hash matches the SHA-1 of the
last line.

## Failure Cases

### File not found

Call the handler with `Path: "nonexistent.go"` and
`Lines: "1-5"`.

Expect: tool error containing `"file not found"`.

### Invalid line range format -- not a range

Call the handler with `Lines: "abc"`.

Expect: tool error containing `"invalid line range"`.

### Invalid line range format -- start greater than end

Call the handler with `Lines: "5-2"`.

Expect: tool error containing `"invalid line range"`.

### Line range out of bounds

Create a file with 3 lines. Call the handler with
`Lines: "1-10"`.

Expect: tool error containing `"invalid line range"` and
the file's actual line count.

### Empty path

Call the handler with `Path: ""` and `Lines: "1-5"`.

Expect: tool error from path validation.

### Path traversal attempt

Call the handler with `Path: "../../etc/passwd"` and
`Lines: "1-5"`.

Expect: tool error from path validation.

### Start line is zero

Call the handler with `Lines: "0-5"`.

Expect: tool error containing `"invalid line range"`.
