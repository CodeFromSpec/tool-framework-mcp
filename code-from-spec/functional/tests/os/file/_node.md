---
depends_on:
  - SPEC/functional/logic/os/file(interface)
output: code-from-spec/functional/tests/os/file/output.md
---

# SPEC/functional/tests/os/file

Test cases for the file operations component.

# Public

## Test cases

### Read mode — happy path

#### Opens and reads all lines

Create a file containing three lines: `"alpha"`, `"beta"`,
`"gamma"` (LF endings). Call `FileOpen` with mode `"read"`,
then call `FileReadLine` three times. Expect `"alpha"`,
`"beta"`, `"gamma"` in order. A fourth `FileReadLine`
raises "end of file".

#### Normalizes CRLF to LF

Create a file containing `"alpha"` and `"beta"` with CRLF
endings. Call `FileOpen` with mode `"read"`, then
`FileReadLine` twice. Expect `"alpha"` and `"beta"` — no
CR or LF characters in the returned strings.

#### Reads file with no trailing newline

Create a file containing `"alpha"` (with LF) and `"beta"`
(no trailing newline). Call `FileOpen` with mode `"read"`,
then `FileReadLine` twice. Expect `"alpha"` and `"beta"`.
A third `FileReadLine` raises EndOfFile.

#### FileSkipLines advances the reader

Create a file containing `"one"`, `"two"`, `"three"`,
`"four"`, `"five"`. Call `FileOpen` with mode `"read"`,
then `FileSkipLines(2)`, then `FileReadLine`. Expect
`"three"`.

#### FileSkipLines past end of file

Create a file containing `"one"`, `"two"`. Call `FileOpen`
with mode `"read"`, then `FileSkipLines(10)`. Expect no
error. Then `FileReadLine` raises EndOfFile.

#### Preserves leading whitespace

Create a file containing `"  alpha"` and `"    beta"`.
Call `FileOpen` with mode `"read"`, then `FileReadLine`
twice. Expect `"  alpha"` and `"    beta"` — leading
spaces preserved.

#### Preserves trailing whitespace

Create a file containing `"alpha  "` and `"beta   "`.
Call `FileOpen` with mode `"read"`, then `FileReadLine`
twice. Expect `"alpha  "` and `"beta   "` — trailing
spaces preserved.

#### Preserves internal whitespace

Create a file containing `"alpha   beta"` and
`"one\ttwo"`. Call `FileOpen` with mode `"read"`, then
`FileReadLine` twice. Expect `"alpha   beta"` and
`"one\ttwo"` — internal spaces and tabs preserved.

#### Preserves empty lines

Create a file containing `"alpha"`, `""`, `""`, `"beta"`.
Call `FileOpen` with mode `"read"`, then `FileReadLine`
four times. Expect `"alpha"`, `""`, `""`, `"beta"` —
empty lines returned as empty strings, not skipped.

#### Preserves non-ASCII characters

Create a file containing `"café"`, `"日本語"`, `"🎉🚀"`.
Call `FileOpen` with mode `"read"`, then `FileReadLine`
three times. Expect `"café"`, `"日本語"`, `"🎉🚀"` —
accented characters, CJK, and emoji pass through
unchanged.

### Read mode — edge cases

#### Empty file

Create an empty file. Call `FileOpen` with mode `"read"`,
then `FileReadLine`. Expect EndOfFile immediately.

#### Single line without newline

Create a file containing only `"hello"` with no newline.
Call `FileOpen` with mode `"read"`, then `FileReadLine`.
Expect `"hello"`. A second `FileReadLine` raises EndOfFile.

### Read mode — failure cases

#### File does not exist

Call `FileOpen` with mode `"read"` and a non-existent
path. Expect FileUnreadable.

#### Read after close

Create a file containing `"alpha"`. Call `FileOpen` with
mode `"read"`, then `FileClose`, then `FileReadLine`.
Expect EndOfFile.

#### Skip after close

Create a file containing `"alpha"`. Call `FileOpen` with
mode `"read"`, then `FileClose`, then `FileSkipLines(1)`.
Expect no error — the call does nothing.

### Overwrite mode — happy path

#### Writes content to a new file

Call `FileOpen` with mode `"overwrite"` for a path that
does not exist, then `FileWrite` with content
`"hello world"`, then `FileClose`. Expect the file to
be created with exactly that content.

#### Overwrites an existing file

Create a file with content `"old"`. Call `FileOpen` with
mode `"overwrite"`, then `FileWrite` with content
`"new"`, then `FileClose`. Expect the file to contain
`"new"` — the old content is replaced entirely.

#### Creates intermediate directories

Call `FileOpen` with mode `"overwrite"` for a path whose
parent directories do not exist (e.g., `"a/b/c/file.txt"`),
then `FileWrite` with content `"data"`, then `FileClose`.
Expect the file and all intermediate directories to be
created.

#### Preserves UTF-8 content

Call `FileOpen` with mode `"overwrite"`, then `FileWrite`
with content containing non-ASCII characters:
`"café 日本語 🎉"`, then `FileClose`. Read the file back.
Expect the content to match byte-for-byte.

#### Preserves line endings as received

Call `FileOpen` with mode `"overwrite"`, then `FileWrite`
with content containing CRLF line endings:
`"alpha\r\nbeta\r\n"`, then `FileClose`. Read the file
back. Expect the content to contain CRLF — no
normalization.

#### Writes empty content

Call `FileOpen` with mode `"overwrite"`, then `FileWrite`
with an empty string, then `FileClose`. Expect a file
to be created with zero bytes.

### Overwrite mode — failure cases

#### Propagates validation errors from PathCfsToOs

Call `FileOpen` with mode `"overwrite"` and an invalid
`PathCfs` (e.g., `"../../outside"`). Expect error
DirectoryTraversal (propagated from PathUtils). Expect no
file or directory to be created.

#### Cannot create directory

Call `FileOpen` with mode `"overwrite"` for a path where
an intermediate directory cannot be created (e.g., a path
component conflicts with an existing file). Expect error
CannotCreateDirectory.

#### Cannot open file (path is a directory)

Create a directory at the target path. Call `FileOpen`
with mode `"overwrite"` pointing to that directory. Expect
error CannotOpenFile.

### Append mode — happy path

#### Append opens without truncating

Create a file with content `"old"`. Call `FileOpen` with
mode `"append"`, then `FileClose` without writing. Read
the file back. Expect `"old"` — content is preserved.

#### Append creates file if it does not exist

Call `FileOpen` with mode `"append"` for a path that does
not exist, then `FileClose`. Expect the file to be created
(empty).

### Wrong mode — failure cases

#### FileReadLine fails in overwrite mode

Call `FileOpen` with mode `"overwrite"`, then
`FileReadLine`. Expect WrongMode error.

#### FileReadLine fails in append mode

Call `FileOpen` with mode `"append"`, then `FileReadLine`.
Expect WrongMode error.

#### FileWrite fails in read mode

Create a file. Call `FileOpen` with mode `"read"`, then
`FileWrite` with any content. Expect WrongMode error.

#### FileSkipLines fails in overwrite mode

Call `FileOpen` with mode `"overwrite"`, then
`FileSkipLines(1)`. Expect WrongMode error.

### Invalid mode — failure case

#### FileOpen rejects unknown mode

Call `FileOpen` with mode `"invalid"`. Expect InvalidMode
error.

### FileRename — happy path

#### Renames a file

Create a file with content `"data"` at path `"a.txt"`.
Call `FileRename` from `"a.txt"` to `"b.txt"`. Expect
`"b.txt"` to exist with content `"data"` and `"a.txt"`
to no longer exist.

#### Rename overwrites destination

Create a file with content `"old"` at `"dest.txt"` and
a file with content `"new"` at `"src.txt"`. Call
`FileRename` from `"src.txt"` to `"dest.txt"`. Expect
`"dest.txt"` to contain `"new"`.

### FileRename — failure cases

#### Rename non-existent source

Call `FileRename` with a source that does not exist.
Expect CannotRename error.

### FileDelete — happy path

#### Deletes a file

Create a file at `"target.txt"`. Call `FileDelete` with
`"target.txt"`. Expect the file to no longer exist.

### FileDelete — failure cases

#### Delete non-existent file

Call `FileDelete` with a path that does not exist. Expect
CannotDelete error.

### Locking — concurrency

#### Shared lock allows concurrent readers

Create a file with content `"data"`. Call `FileOpen` with
mode `"read"` to get handle1. In a separate goroutine,
call `FileOpen` with mode `"read"` on the same file to
get handle2. Expect handle2 to open successfully (shared
lock does not block other shared locks). Close both
handles.

#### Exclusive lock blocks other exclusive locks

Create a file with content `"data"`. Call `FileOpen` with
mode `"overwrite"` to get handle1 (exclusive lock). In a
separate goroutine, call `FileOpen` with mode `"overwrite"`
on the same file. Expect the second open to block (not
return immediately). Close handle1. Expect the second open
to succeed after handle1 is closed. Close both handles.

#### Exclusive lock blocks shared locks

Create a file with content `"data"`. Call `FileOpen` with
mode `"overwrite"` to get handle1 (exclusive lock). In a
separate goroutine, call `FileOpen` with mode `"read"` on
the same file. Expect the read open to block. Close
handle1. Expect the read open to succeed. Close both
handles.

#### Append mode acquires exclusive lock

Create a file with content `"data"`. Call `FileOpen` with
mode `"append"` to get handle1 (exclusive lock). In a
separate goroutine, call `FileOpen` with mode `"read"` on
the same file. Expect the read open to block. Close
handle1. Expect the read open to succeed. Close both
handles.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function names from the interface: `FileOpen`,
  `FileReadLine`, `FileWrite`, `FileSkipLines`, `FileClose`,
  `FileRename`, `FileDelete`.
