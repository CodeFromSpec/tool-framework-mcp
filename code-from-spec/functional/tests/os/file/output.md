<!-- code-from-spec: SPEC/functional/tests/os/file@pQ9xybyGeL94JKgDoQkhAgVsqp4 -->

## Test: Opens and reads all lines

Setup: a file containing three lines `"alpha"`, `"beta"`, `"gamma"` with LF endings.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect `"gamma"`.
5. Call `FileReadLine`. Expect EndOfFile.

---

## Test: Normalizes CRLF to LF

Setup: a file containing `"alpha"` and `"beta"` with CRLF line endings.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"alpha"` — no CR character.
3. Call `FileReadLine`. Expect `"beta"` — no CR character.

---

## Test: Reads file with no trailing newline

Setup: a file containing `"alpha"` followed by LF, then `"beta"` with no trailing newline.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect EndOfFile.

---

## Test: FileSkipLines advances the reader

Setup: a file containing five lines: `"one"`, `"two"`, `"three"`, `"four"`, `"five"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileSkipLines` with count 2. Expect no error.
3. Call `FileReadLine`. Expect `"three"`.

---

## Test: FileSkipLines past end of file

Setup: a file containing two lines: `"one"`, `"two"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileSkipLines` with count 10. Expect no error.
3. Call `FileReadLine`. Expect EndOfFile.

---

## Test: Preserves leading whitespace

Setup: a file containing `"  alpha"` and `"    beta"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"  alpha"` — two leading spaces intact.
3. Call `FileReadLine`. Expect `"    beta"` — four leading spaces intact.

---

## Test: Preserves trailing whitespace

Setup: a file containing `"alpha  "` and `"beta   "`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"alpha  "` — two trailing spaces intact.
3. Call `FileReadLine`. Expect `"beta   "` — three trailing spaces intact.

---

## Test: Preserves internal whitespace

Setup: a file containing `"alpha   beta"` and `"one\ttwo"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"alpha   beta"` — internal spaces intact.
3. Call `FileReadLine`. Expect `"one\ttwo"` — tab character intact.

---

## Test: Preserves empty lines

Setup: a file containing four lines: `"alpha"`, `""`, `""`, `"beta"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `""` — empty string, not skipped.
4. Call `FileReadLine`. Expect `""` — empty string, not skipped.
5. Call `FileReadLine`. Expect `"beta"`.

---

## Test: Preserves non-ASCII characters

Setup: a file containing `"café"`, `"日本語"`, `"🎉🚀"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"café"` — accented characters intact.
3. Call `FileReadLine`. Expect `"日本語"` — CJK characters intact.
4. Call `FileReadLine`. Expect `"🎉🚀"` — emoji intact.

---

## Test: Empty file

Setup: an empty file (zero bytes).

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect EndOfFile immediately.

---

## Test: Single line without newline

Setup: a file containing only `"hello"` with no newline character.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect `"hello"`.
3. Call `FileReadLine`. Expect EndOfFile.

---

## Test: File does not exist

Setup: a path that points to no existing file.

Actions:
1. Call `FileOpen` with the non-existent path and mode `"read"`. Expect FileUnreadable.

---

## Test: Read after close

Setup: a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileClose` on the handle.
3. Call `FileReadLine` on the closed handle. Expect EndOfFile.

---

## Test: Skip after close

Setup: a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileClose` on the handle.
3. Call `FileSkipLines` with count 1 on the closed handle. Expect no error — the call does nothing.

---

## Test: Writes content to a new file

Setup: a path that does not exist.

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileWrite` with content `"hello world"`.
3. Call `FileClose`.
4. Read the file back. Expect the content to be exactly `"hello world"`.

---

## Test: Overwrites an existing file

Setup: a file with content `"old"`.

Actions:
1. Call `FileOpen` with the file path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileWrite` with content `"new"`.
3. Call `FileClose`.
4. Read the file back. Expect the content to be `"new"` — old content entirely replaced.

---

## Test: Creates intermediate directories

Setup: a path whose parent directories do not exist (e.g., `"a/b/c/file.txt"`).

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileWrite` with content `"data"`.
3. Call `FileClose`.
4. Expect the file and all intermediate directories to be created with content `"data"`.

---

## Test: Preserves UTF-8 content

Setup: no existing file at the target path.

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileWrite` with content `"café 日本語 🎉"`.
3. Call `FileClose`.
4. Read the file back. Expect the content to match byte-for-byte.

---

## Test: Preserves line endings as received

Setup: no existing file at the target path.

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileWrite` with content `"alpha\r\nbeta\r\n"` (CRLF endings).
3. Call `FileClose`.
4. Read the file back. Expect the content to contain CRLF — no normalization applied.

---

## Test: Writes empty content

Setup: no existing file at the target path.

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileWrite` with an empty string.
3. Call `FileClose`.
4. Expect a file to be created with zero bytes.

---

## Test: Propagates validation errors from PathCfsToOs

Setup: an invalid PathCfs value such as `"../../outside"`.

Actions:
1. Call `FileOpen` with the invalid path and mode `"overwrite"`. Expect DirectoryTraversal (propagated from PathUtils).
2. Expect no file or directory to be created.

---

## Test: Cannot create directory

Setup: a path where an intermediate component conflicts with an existing file (e.g., a file exists at a path that must be a directory).

Actions:
1. Call `FileOpen` with the conflicting path and mode `"overwrite"`. Expect CannotCreateDirectory.

---

## Test: Cannot open file (path is a directory)

Setup: a directory exists at the target path.

Actions:
1. Call `FileOpen` pointing to that directory with mode `"overwrite"`. Expect CannotOpenFile.

---

## Test: Append opens without truncating

Setup: a file with content `"old"`.

Actions:
1. Call `FileOpen` with the file path and mode `"append"`. Expect a FileHandle.
2. Call `FileClose` without writing.
3. Read the file back. Expect `"old"` — content is preserved.

---

## Test: Append creates file if it does not exist

Setup: a path that does not exist.

Actions:
1. Call `FileOpen` with the path and mode `"append"`. Expect a FileHandle.
2. Call `FileClose`.
3. Expect the file to be created (empty).

---

## Test: FileReadLine fails in overwrite mode

Setup: no existing file at the target path.

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect WrongMode.

---

## Test: FileReadLine fails in append mode

Setup: no existing file at the target path.

Actions:
1. Call `FileOpen` with the path and mode `"append"`. Expect a FileHandle.
2. Call `FileReadLine`. Expect WrongMode.

---

## Test: FileWrite fails in read mode

Setup: an existing file.

Actions:
1. Call `FileOpen` with the file path and mode `"read"`. Expect a FileHandle.
2. Call `FileWrite` with any content. Expect WrongMode.

---

## Test: FileSkipLines fails in overwrite mode

Setup: no existing file at the target path.

Actions:
1. Call `FileOpen` with the path and mode `"overwrite"`. Expect a FileHandle.
2. Call `FileSkipLines` with count 1. Expect WrongMode.

---

## Test: FileOpen rejects unknown mode

Setup: none.

Actions:
1. Call `FileOpen` with any path and mode `"invalid"`. Expect InvalidMode.

---

## Test: Renames a file

Setup: a file with content `"data"` at path `"a.txt"`.

Actions:
1. Call `FileRename` with source `"a.txt"` and destination `"b.txt"`.
2. Expect `"b.txt"` to exist with content `"data"`.
3. Expect `"a.txt"` to no longer exist.

---

## Test: Rename overwrites destination

Setup: a file with content `"old"` at `"dest.txt"` and a file with content `"new"` at `"src.txt"`.

Actions:
1. Call `FileRename` with source `"src.txt"` and destination `"dest.txt"`.
2. Expect `"dest.txt"` to contain `"new"`.

---

## Test: Rename non-existent source

Setup: a source path that does not exist.

Actions:
1. Call `FileRename` with the non-existent source and any destination. Expect CannotRename.

---

## Test: Deletes a file

Setup: a file at `"target.txt"`.

Actions:
1. Call `FileDelete` with `"target.txt"`.
2. Expect the file to no longer exist.

---

## Test: Delete non-existent file

Setup: a path that does not exist.

Actions:
1. Call `FileDelete` with the non-existent path. Expect CannotDelete.
