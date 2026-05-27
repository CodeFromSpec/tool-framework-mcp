---
depends_on:
  - ROOT/functional/logic/mcp_tools/hash_fragment(interface)
outputs:
  - id: hash_fragment_tests
    path: code-from-spec/functional/tests/mcp_tools/hash_fragment/output.md
---

# ROOT/functional/tests/mcp_tools/hash_fragment

Test cases for the hash fragment tool.

# Public

## Test cases

### Happy path

#### Hashes a valid line range

Create a file with multiple lines of known content. Call
the hash fragment handler with path pointing to the file
and lines = "2-4".

Expect success. The result contains a 27-character
base64url-encoded SHA-1 hash matching the expected SHA-1
of lines 2 through 4 (inclusive, joined with LF).

#### Single line range

Create a file with multiple lines. Call the handler with
lines = "3-3".

Expect success. The hash matches the SHA-1 of just line 3.

#### First line of file

Create a file with multiple lines. Call the handler with
lines = "1-1".

Expect success. The hash matches the SHA-1 of the first
line only.

#### Last line of file

Create a file with exactly 5 lines. Call the handler with
lines = "5-5".

Expect success. The hash matches the SHA-1 of the last
line.

### Failure cases

#### File not found

Call the handler with path = "nonexistent.go" and
lines = "1-5". Expect error containing "file not found".

#### Invalid line range format -- not a range

Call the handler with lines = "abc". Expect error
containing "invalid line range".

#### Invalid line range format -- start greater than end

Call the handler with lines = "5-2". Expect error
containing "invalid line range".

#### Line range out of bounds

Create a file with 3 lines. Call the handler with
lines = "1-10". Expect error containing "invalid line
range" and the file's actual line count.

#### Empty path

Call the handler with path = "" and lines = "1-5". Expect
error from path validation.

#### Path traversal attempt

Call the handler with path = "../../etc/passwd" and
lines = "1-5". Expect error from path validation.

#### Start line is zero

Call the handler with lines = "0-5". Expect error
containing "invalid line range".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
