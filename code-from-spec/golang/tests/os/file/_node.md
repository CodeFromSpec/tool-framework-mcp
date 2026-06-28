---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
output: internal/file/file_test.go
---

# SPEC/golang/tests/os/file

# Agent

## Test cases

### Read mode — happy path

#### Opens and reads all lines

Setup: file with "alpha", "beta", "gamma" (LF).

Actions: FileOpen("read"), FileReadLine x3, then x4.

Expected: "alpha", "beta", "gamma", then EndOfFile.

#### Normalizes CRLF to LF

Setup: file with "alpha", "beta" (CRLF).

Expected: "alpha", "beta" — no CR characters.

#### Reads file with no trailing newline

Setup: "alpha\nbeta" (no trailing newline).

Expected: "alpha", "beta", then EndOfFile.

#### FileSkipLines advances the reader

Setup: file with 5 lines. FileSkipLines(2).

Expected: FileReadLine returns "three".

#### FileSkipLines past end of file

Setup: file with 2 lines. FileSkipLines(10).

Expected: No error. FileReadLine raises EndOfFile.

#### Preserves leading whitespace

Setup: "  alpha", "    beta".

Expected: Leading spaces preserved.

#### Preserves trailing whitespace

Setup: "alpha  ", "beta   ".

Expected: Trailing spaces preserved.

#### Preserves internal whitespace

Setup: "alpha   beta", "one\ttwo".

Expected: Internal spaces and tabs preserved.

#### Preserves empty lines

Setup: "alpha", "", "", "beta".

Expected: Empty lines as empty strings, not skipped.

#### Preserves non-ASCII characters

Setup: "café", "日本語", "🎉🚀".

Expected: All characters pass through unchanged.

### Read mode — edge cases

#### Empty file

Setup: empty file (zero bytes).

Expected: FileReadLine raises EndOfFile immediately.

#### Single line without newline

Setup: "hello" with no newline.

Expected: "hello", then EndOfFile.

### Read mode — failure cases

#### File does not exist

Expected: FileOpen raises ErrFileUnreadable.

#### Read after close

Setup: file with "alpha". FileOpen, FileClose.

Expected: FileReadLine raises ErrEndOfFile.

#### Skip after close

Setup: file with "alpha". FileOpen, FileClose.

Expected: FileSkipLines(1) — no error, does nothing.

### Overwrite mode — happy path

#### Writes content to a new file

Expected: file created with "hello world".

#### Overwrites an existing file

Setup: file with "old". FileOpen("overwrite"),
FileWrite("new").

Expected: file contains "new".

#### Creates intermediate directories

Setup: path "a/b/c/file.txt" (dirs don't exist).

Expected: file and all intermediate dirs created.

#### Preserves UTF-8 content

FileWrite("café 日本語 🎉"). Read back.

Expected: byte-for-byte match.

#### Preserves line endings as received

FileWrite("alpha\r\nbeta\r\n"). Read back.

Expected: CRLF preserved — no normalization.

#### Writes empty content

FileWrite(""). Expected: file with zero bytes.

### Overwrite mode — failure cases

#### Propagates validation errors from PathCfsToOs

PathCfs "../../outside". Expected:
ErrDirectoryTraversal. No file created.

#### Cannot create directory

Path component conflicts with existing file.

Expected: ErrCannotCreateDirectory.

#### Cannot open file (path is a directory)

Directory exists at target path.

Expected: ErrCannotOpenFile.

### Append mode — happy path

#### Append opens without truncating

Setup: file with "old". FileOpen("append"), FileClose.

Expected: file still contains "old".

#### Append creates file if it does not exist

FileOpen("append") for non-existent path, FileClose.

Expected: file created (empty).

### Wrong mode — failure cases

#### FileReadLine fails in overwrite mode

Expected: ErrWrongMode.

#### FileReadLine fails in append mode

Expected: ErrWrongMode.

#### FileWrite fails in read mode

Expected: ErrWrongMode.

#### FileSkipLines fails in overwrite mode

Expected: ErrWrongMode.

### Invalid mode — failure case

#### FileOpen rejects unknown mode

Mode "invalid". Expected: ErrInvalidMode.

### FileRename — happy path

#### Renames a file

"a.txt" with "data" → "b.txt". Expected: "b.txt"
exists with "data", "a.txt" gone.

#### Rename overwrites destination

"dest.txt" with "old", "src.txt" with "new".
FileRename("src.txt", "dest.txt").

Expected: "dest.txt" contains "new".

### FileRename — failure cases

#### Rename non-existent source

Expected: ErrCannotRename.

### FileDelete — happy path

#### Deletes a file

"target.txt" exists. FileDelete. Expected: gone.

### FileDelete — failure cases

#### Delete non-existent file

Expected: ErrCannotDelete.

### Locking — concurrency

#### Shared lock allows concurrent readers

Two FileOpen("read") on same file. Both succeed.

#### Exclusive lock blocks other exclusive locks

FileOpen("overwrite") holds lock. Second
FileOpen("overwrite") blocks until first closes.

#### Exclusive lock blocks shared locks

FileOpen("overwrite") holds lock.
FileOpen("read") blocks until first closes.

#### Append mode acquires exclusive lock

FileOpen("append") holds lock.
FileOpen("read") blocks until first closes.

## Go-specific guidance

- The package name is `file_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- For concurrency tests, use goroutines with channels
  for synchronization.
