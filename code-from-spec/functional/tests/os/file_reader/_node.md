---
depends_on:
  - SPEC/functional/logic/os/file_reader(interface)
output: code-from-spec/functional/tests/os/file_reader/output.md
---

# SPEC/functional/tests/os/file_reader

Test cases for the file reader component.

# Public

## Test cases

### Happy path

#### Opens and reads all lines

Create a file containing three lines: `"alpha"`, `"beta"`,
`"gamma"` (LF endings). Call `FileOpen`, then call
`FileReadLine` three times. Expect `"alpha"`, `"beta"`,
`"gamma"` in order. A fourth `FileReadLine` raises
"end of file".

#### Normalizes CRLF to LF

Create a file containing `"alpha"` and `"beta"` with CRLF
endings. Call `FileOpen`, then `FileReadLine` twice.
Expect `"alpha"` and `"beta"` — no CR or LF characters
in the returned strings.

#### Reads file with no trailing newline

Create a file containing `"alpha"` (with LF) and `"beta"`
(no trailing newline). Call `FileOpen`, then `FileReadLine`
twice. Expect `"alpha"` and `"beta"`. A third
`FileReadLine` raises EndOfFile.

#### FileSkipLines advances the reader

Create a file containing `"one"`, `"two"`, `"three"`,
`"four"`, `"five"`. Call `FileOpen`, then
`FileSkipLines(2)`, then `FileReadLine`. Expect `"three"`.

#### FileSkipLines past end of file

Create a file containing `"one"`, `"two"`. Call `FileOpen`,
then `FileSkipLines(10)`. Expect no error. Then
`FileReadLine` raises EndOfFile.

#### Preserves leading whitespace

Create a file containing `"  alpha"` and `"    beta"`.
Call `FileOpen`, then `FileReadLine` twice. Expect
`"  alpha"` and `"    beta"` — leading spaces preserved.

#### Preserves trailing whitespace

Create a file containing `"alpha  "` and `"beta   "`.
Call `FileOpen`, then `FileReadLine` twice. Expect
`"alpha  "` and `"beta   "` — trailing spaces preserved.

#### Preserves internal whitespace

Create a file containing `"alpha   beta"` and
`"one\ttwo"`. Call `FileOpen`, then `FileReadLine` twice.
Expect `"alpha   beta"` and `"one\ttwo"` — internal
spaces and tabs preserved.

#### Preserves empty lines

Create a file containing `"alpha"`, `""`, `""`, `"beta"`.
Call `FileOpen`, then `FileReadLine` four times. Expect
`"alpha"`, `""`, `""`, `"beta"` — empty lines returned
as empty strings, not skipped.

#### Preserves non-ASCII characters

Create a file containing `"café"`, `"日本語"`, `"🎉🚀"`.
Call `FileOpen`, then `FileReadLine` three times. Expect
`"café"`, `"日本語"`, `"🎉🚀"` — accented characters,
CJK, and emoji pass through unchanged.

### Edge cases

#### Empty file

Create an empty file. Call `FileOpen`, then `FileReadLine`.
Expect EndOfFile immediately.

#### Single line without newline

Create a file containing only `"hello"` with no newline.
Call `FileOpen`, then `FileReadLine`. Expect `"hello"`.
A second `FileReadLine` raises EndOfFile.

### Failure cases

#### File does not exist

Call `FileOpen` with a non-existent path. Expect
FileUnreadable.

#### Read after close

Create a file containing `"alpha"`. Call `FileOpen`, then
`FileClose`, then `FileReadLine`. Expect EndOfFile.

#### Skip after close

Create a file containing `"alpha"`. Call `FileOpen`, then
`FileClose`, then `FileSkipLines(1)`. Expect no error —
the call does nothing.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function names from the interface: `FileOpen`,
  `FileReadLine`, `FileSkipLines`, `FileClose`.
