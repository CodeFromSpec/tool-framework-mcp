<!-- code-from-spec: ROOT/functional/tests/os/file_reader@pbsSF5a0NBCBipNvBEQNN-74H0k -->

## Test: Opens and reads all lines

Setup: create a file containing three lines `"alpha"`, `"beta"`, `"gamma"` with LF endings.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect `"gamma"`.
5. Call `FileReadLine`. Expect EndOfFile.

---

## Test: Normalizes CRLF to LF

Setup: create a file containing `"alpha"` and `"beta"` with CRLF endings.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"` — no CR or LF in the returned string.
3. Call `FileReadLine`. Expect `"beta"` — no CR or LF in the returned string.

---

## Test: Reads file with no trailing newline

Setup: create a file containing `"alpha"` followed by LF, then `"beta"` with no trailing newline.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect EndOfFile.

---

## Test: FileSkipLines advances the reader

Setup: create a file containing five lines: `"one"`, `"two"`, `"three"`, `"four"`, `"five"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileSkipLines` with count `2`. Expect no error.
3. Call `FileReadLine`. Expect `"three"`.

---

## Test: FileSkipLines past end of file

Setup: create a file containing two lines: `"one"`, `"two"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileSkipLines` with count `10`. Expect no error.
3. Call `FileReadLine`. Expect EndOfFile.

---

## Test: Preserves leading whitespace

Setup: create a file containing two lines: `"  alpha"` and `"    beta"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"  alpha"` — leading spaces preserved.
3. Call `FileReadLine`. Expect `"    beta"` — leading spaces preserved.

---

## Test: Preserves trailing whitespace

Setup: create a file containing two lines: `"alpha  "` and `"beta   "`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha  "` — trailing spaces preserved.
3. Call `FileReadLine`. Expect `"beta   "` — trailing spaces preserved.

---

## Test: Preserves internal whitespace

Setup: create a file containing two lines: `"alpha   beta"` and `"one\ttwo"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha   beta"` — internal spaces preserved.
3. Call `FileReadLine`. Expect `"one\ttwo"` — internal tab preserved.

---

## Test: Preserves empty lines

Setup: create a file containing four lines: `"alpha"`, `""`, `""`, `"beta"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `""` — empty line returned as empty string, not skipped.
4. Call `FileReadLine`. Expect `""` — empty line returned as empty string, not skipped.
5. Call `FileReadLine`. Expect `"beta"`.

---

## Test: Preserves non-ASCII characters

Setup: create a file containing three lines: `"café"`, `"日本語"`, `"🎉🚀"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"café"` — accented characters unchanged.
3. Call `FileReadLine`. Expect `"日本語"` — CJK characters unchanged.
4. Call `FileReadLine`. Expect `"🎉🚀"` — emoji unchanged.

---

## Test: Empty file

Setup: create an empty file.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect EndOfFile immediately.

---

## Test: Single line without newline

Setup: create a file containing only `"hello"` with no newline.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine`. Expect `"hello"`.
3. Call `FileReadLine`. Expect EndOfFile.

---

## Test: File does not exist

Setup: no file is created; use a path that does not exist.

Actions:
1. Call `FileOpen` with the non-existent path. Expect FileUnreadable.

---

## Test: Read after close

Setup: create a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileClose` on the reader. Expect no error.
3. Call `FileReadLine` on the same reader. Expect EndOfFile.

---

## Test: Skip after close

Setup: create a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileClose` on the reader. Expect no error.
3. Call `FileSkipLines` with count `1` on the same reader. Expect no error — the call does nothing.
