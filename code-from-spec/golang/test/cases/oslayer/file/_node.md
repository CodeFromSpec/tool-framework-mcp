---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
output: internal/oslayerfiletest/oslayer_file_test.go
---

# SPEC/golang/test/cases/oslayer/file

# Agent

## Test cases

### Read mode — happy path

#### Opens and reads all lines

Setup: file with "alpha", "beta", "gamma" (LF).

Actions: OpenFile("read"), ReadLine x3, then x4.

Expected: "alpha", "beta", "gamma", then ErrEndOfFile.

#### Normalizes CRLF to LF

Setup: file with "alpha", "beta" (CRLF).

Expected: "alpha", "beta" — no CR characters.

#### Reads file with no trailing newline

Setup: "alpha\nbeta" (no trailing newline).

Expected: "alpha", "beta", then ErrEndOfFile.

#### SkipLines advances the reader

Setup: file with 5 lines. SkipLines(2).

Expected: ReadLine returns "three".

#### SkipLines past end of file

Setup: file with 2 lines. SkipLines(10).

Expected: No error. ReadLine raises ErrEndOfFile.

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

Expected: ReadLine raises ErrEndOfFile immediately.

#### Single line without newline

Setup: "hello" with no newline.

Expected: "hello", then ErrEndOfFile.

### Read mode — failure cases

#### File does not exist

Expected: OpenFile raises ErrFileUnreadable.

#### Read after close

Setup: file with "alpha". OpenFile, Close.

Expected: ReadLine raises ErrEndOfFile.

#### Skip after close

Setup: file with "alpha". OpenFile, Close.

Expected: SkipLines(1) — no error, does nothing.

### Overwrite mode — happy path

#### Writes content to a new file

Expected: file created with "hello world".

#### Overwrites an existing file

Setup: file with "old". OpenFile("overwrite"),
Write("new").

Expected: file contains "new".

#### Creates intermediate directories

Setup: path "a/b/c/file.txt" (dirs don't exist).

Expected: file and all intermediate dirs created.

#### Preserves UTF-8 content

Write("café 日本語 🎉"). Read back.

Expected: byte-for-byte match.

#### Preserves line endings as received

Write("alpha\r\nbeta\r\n"). Read back.

Expected: CRLF preserved — no normalization.

#### Writes empty content

Write(""). Expected: file with zero bytes.

### Overwrite mode — failure cases

#### Propagates validation errors from CfsPathToOs

CfsPath "../../outside". Expected:
ErrDirectoryTraversal. No file created.

#### Cannot create directory

Path component conflicts with existing file.

Expected: ErrCannotCreateDirectory.

#### Cannot open file (path is a directory)

Directory exists at target path.

Expected: ErrCannotOpenFile.

### Append mode — happy path

#### Append opens without truncating

Setup: file with "old". OpenFile("append"), Close.

Expected: file still contains "old".

#### Append creates file if it does not exist

OpenFile("append") for non-existent path, Close.

Expected: file created (empty).

#### Write succeeds in append mode

OpenFile("append"), Write("content"), Close.

Expected: no error, file contains "content".

#### Append actually appends content

Setup: file with "old\n". OpenFile("append"),
Write("new\n"), Close. Read back.

Expected: file contains "old\nnew\n".

#### Append creates intermediate directories

Path "x/y/z/file.txt" (dirs don't exist).
OpenFile("append"), Close.

Expected: file and all intermediate dirs created.

### Wrong mode — failure cases

#### ReadLine fails in overwrite mode

Expected: ErrWrongMode.

#### ReadLine fails in append mode

Expected: ErrWrongMode.

#### Write fails in read mode

Expected: ErrWrongMode.

#### SkipLines fails in overwrite mode

Expected: ErrWrongMode.

#### SkipLines fails in append mode

Expected: ErrWrongMode.

### Invalid mode — failure case

#### OpenFile rejects unknown mode

Mode "invalid". Expected: ErrInvalidMode.

### RenameFile — happy path

#### Renames a file

"a.txt" with "data" → "b.txt". Expected: "b.txt"
exists with "data", "a.txt" gone.

#### Rename overwrites destination

"dest.txt" with "old", "src.txt" with "new".
RenameFile("src.txt", "dest.txt").

Expected: "dest.txt" contains "new".

### RenameFile — failure cases

#### Rename non-existent source

Expected: ErrCannotRename.

#### Rename with invalid CfsPath

Source: "../../outside". Expected: validation error
propagated (ErrDirectoryTraversal).

### DeleteFile — happy path

#### Deletes a file

"target.txt" exists. DeleteFile. Expected: gone.

### DeleteFile — failure cases

#### Delete non-existent file

Expected: ErrCannotDelete.

#### Delete with invalid CfsPath

Input: "../../outside". Expected: validation error
propagated (ErrDirectoryTraversal).

### Locking — concurrency

#### Shared lock allows concurrent readers

Two OpenFile("read") on same file. Both succeed.

#### Exclusive lock blocks other exclusive locks

OpenFile("overwrite") holds lock. Second
OpenFile("overwrite") blocks until first closes.

#### Exclusive lock blocks shared locks

OpenFile("overwrite") holds lock.
OpenFile("read") blocks until first closes.

#### Append mode acquires exclusive lock

OpenFile("append") holds lock.
OpenFile("read") blocks until first closes.

#### Lock timeout

OpenFile("overwrite") holds lock. Second
OpenFile("overwrite") with short timeoutMs.

Expected: ErrLockTimeout.

## Go-specific guidance

- The package name is `oslayerfiletest` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- For concurrency tests, use goroutines with channels
  for synchronization.
