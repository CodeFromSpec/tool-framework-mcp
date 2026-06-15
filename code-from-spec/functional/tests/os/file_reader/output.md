<!-- code-from-spec: SPEC/functional/tests/os/file_reader@U7MJe2Ijzb9YcjLJ_sdt4EE5f74 -->

## Test: Opens and reads all lines

Setup: a file containing three lines `"alpha"`, `"beta"`, `"gamma"` with LF endings.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect `"gamma"`.
5. Call `FileReadLine`. Expect EndOfFile.

---

## Test: Normalizes CRLF to LF

Setup: a file containing `"alpha"` and `"beta"` with CRLF line endings.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"` — no CR character.
3. Call `FileReadLine`. Expect `"beta"` — no CR character.

---

## Test: Reads file with no trailing newline

Setup: a file containing `"alpha"` followed by LF, then `"beta"` with no trailing newline.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect EndOfFile.

---

## Test: FileSkipLines advances the reader

Setup: a file containing five lines: `"one"`, `"two"`, `"three"`, `"four"`, `"five"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileSkipLines` with count 2. Expect no error.
3. Call `FileReadLine`. Expect `"three"`.

---

## Test: FileSkipLines past end of file

Setup: a file containing two lines: `"one"`, `"two"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileSkipLines` with count 10. Expect no error.
3. Call `FileReadLine`. Expect EndOfFile.

---

## Test: Preserves leading whitespace

Setup: a file containing `"  alpha"` and `"    beta"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"  alpha"` — two leading spaces intact.
3. Call `FileReadLine`. Expect `"    beta"` — four leading spaces intact.

---

## Test: Preserves trailing whitespace

Setup: a file containing `"alpha  "` and `"beta   "`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha  "` — two trailing spaces intact.
3. Call `FileReadLine`. Expect `"beta   "` — three trailing spaces intact.

---

## Test: Preserves internal whitespace

Setup: a file containing `"alpha   beta"` and `"one\ttwo"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha   beta"` — internal spaces intact.
3. Call `FileReadLine`. Expect `"one\ttwo"` — tab character intact.

---

## Test: Preserves empty lines

Setup: a file containing four lines: `"alpha"`, `""`, `""`, `"beta"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `""` — empty string, not skipped.
4. Call `FileReadLine`. Expect `""` — empty string, not skipped.
5. Call `FileReadLine`. Expect `"beta"`.

---

## Test: Preserves non-ASCII characters

Setup: a file containing `"café"`, `"日本語"`, `"🎉🚀"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"café"` — accented characters intact.
3. Call `FileReadLine`. Expect `"日本語"` — CJK characters intact.
4. Call `FileReadLine`. Expect `"🎉🚀"` — emoji intact.

---

## Test: Empty file

Setup: an empty file (zero bytes).

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect EndOfFile immediately.

---

## Test: Single line without newline

Setup: a file containing only `"hello"` with no newline character.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"hello"`.
3. Call `FileReadLine`. Expect EndOfFile.

---

## Test: File does not exist

Setup: a path that points to no existing file.

Actions:
1. Call `FileOpen` with the non-existent path. Expect FileUnreadable.

---

## Test: Read after close

Setup: a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileClose` on the reader.
3. Call `FileReadLine` on the closed reader. Expect EndOfFile.

---

## Test: Skip after close

Setup: a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileClose` on the reader.
3. Call `FileSkipLines` with count 1 on the closed reader. Expect no error — the call does nothing.
