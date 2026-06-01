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

Fragment hashes use SHA-1 encoded as base64url (RFC 4648
§5, no padding) — always 27 characters. The input to
SHA-1 is the extracted lines, each with `\n` appended
(including the last line). Tests that verify a specific
hash value must compute it using this algorithm.

### Happy path

#### Hashes a valid line range

Create a file with 5 lines of known content. Call
MCPHashFragment with path pointing to the file and
lines = "2-4".

Expect success. The result is a 27-character string
matching the SHA-1 of lines 2-4 (each with `\n`
appended).

#### Single line range

Create a file with 5 lines. Call MCPHashFragment with
lines = "3-3".

Expect success. The hash matches the SHA-1 of line 3
with `\n` appended.

#### First line of file

Create a file with 5 lines. Call MCPHashFragment with
lines = "1-1".

Expect success. The hash matches the SHA-1 of the
first line with `\n` appended.

#### Last line of file

Create a file with exactly 5 lines. Call MCPHashFragment
with lines = "5-5".

Expect success. The hash matches the SHA-1 of the last
line with `\n` appended.

#### Hash is deterministic

Create a file with known content. Call MCPHashFragment
twice with the same path and lines. Expect both results
are identical.

### Error cases

#### File does not exist

Call MCPHashFragment with path = "nonexistent.go" and
lines = "1-5". Expect error FileUnreadable (propagated
from FileReader via FileOpen).

#### Invalid line range format — not a range

Call MCPHashFragment with lines = "abc". Expect error
InvalidLineRange.

#### Start greater than end

Call MCPHashFragment with lines = "5-2". Expect error
InvalidLineRange.

#### Start less than 1

Call MCPHashFragment with lines = "0-5". Expect error
InvalidLineRange.

#### Line range out of bounds

Create a file with 3 lines. Call MCPHashFragment with
lines = "1-10". Expect error InvalidLineRange.

#### Empty path

Call MCPHashFragment with path = "" and lines = "1-5".
Expect error PathEmpty (propagated from PathUtils via
PathValidateCfs).

#### Path traversal

Call MCPHashFragment with path = "../../etc/passwd" and
lines = "1-5". Expect error DirectoryTraversal
(propagated from PathUtils via PathValidateCfs).

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `MCPHashFragment`.
- Use formal error names (PascalCase) as defined in the
  interface.
- When verifying hash values, compute the expected hash
  per the hashing convention described above.
