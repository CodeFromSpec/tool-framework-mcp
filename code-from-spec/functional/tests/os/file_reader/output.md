<!-- code-from-spec: ROOT/functional/tests/os/file_reader@cZx6V8X4VKJaJpAze78a-0Cy09o -->

# FileReader Test Specification

## Happy Path

### Opens and reads all lines

Setup: Create a file containing three lines — `"alpha"`, `"beta"`, `"gamma"` — with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"`.
3. Call `FileReadLine` — expect `"beta"`.
4. Call `FileReadLine` — expect `"gamma"`.
5. Call `FileReadLine` — expect EndOfFile.

---

### Normalizes CRLF to LF

Setup: Create a file containing `"alpha"` and `"beta"` with CRLF line endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"` (no CR or LF in the returned string).
3. Call `FileReadLine` — expect `"beta"` (no CR or LF in the returned string).

---

### Reads file with no trailing newline

Setup: Create a file containing `"alpha"` (with LF) followed by `"beta"` (no trailing newline).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"`.
3. Call `FileReadLine` — expect `"beta"`.
4. Call `FileReadLine` — expect EndOfFile.

---

### FileSkipLines advances the reader

Setup: Create a file containing five lines — `"one"`, `"two"`, `"three"`, `"four"`, `"five"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileSkipLines` with count `2`.
3. Call `FileReadLine` — expect `"three"`.

---

### FileSkipLines past end of file

Setup: Create a file containing two lines — `"one"`, `"two"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileSkipLines` with count `10` — expect no error.
3. Call `FileReadLine` — expect EndOfFile.

---

### Preserves leading whitespace

Setup: Create a file containing `"  alpha"` and `"    beta"` (with leading spaces).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"  alpha"` (two leading spaces preserved).
3. Call `FileReadLine` — expect `"    beta"` (four leading spaces preserved).

---

### Preserves trailing whitespace

Setup: Create a file containing `"alpha  "` and `"beta   "` (with trailing spaces).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha  "` (two trailing spaces preserved).
3. Call `FileReadLine` — expect `"beta   "` (three trailing spaces preserved).

---

### Preserves internal whitespace

Setup: Create a file containing `"alpha   beta"` (internal spaces) and `"one\ttwo"` (tab character).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha   beta"` (internal spaces preserved).
3. Call `FileReadLine` — expect `"one\ttwo"` (tab preserved).

---

### Preserves empty lines

Setup: Create a file containing four lines — `"alpha"`, `""`, `""`, `"beta"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"`.
3. Call `FileReadLine` — expect `""` (empty string, not skipped).
4. Call `FileReadLine` — expect `""` (empty string, not skipped).
5. Call `FileReadLine` — expect `"beta"`.

---

### Preserves non-ASCII characters

Setup: Create a file containing `"café"`, `"日本語"`, `"🎉🚀"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"café"` (accented characters unchanged).
3. Call `FileReadLine` — expect `"日本語"` (CJK characters unchanged).
4. Call `FileReadLine` — expect `"🎉🚀"` (emoji unchanged).

---

## Edge Cases

### Empty file

Setup: Create an empty file (zero bytes).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect EndOfFile immediately.

---

### Single line without newline

Setup: Create a file containing only `"hello"` with no newline character.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"hello"`.
3. Call `FileReadLine` — expect EndOfFile.

---

## Failure Cases

### File does not exist

Setup: No file is created; use a path that does not exist on the filesystem.

Actions:
1. Call `FileOpen` with the non-existent path — expect FileUnreadable.

---

### Read after close

Setup: Create a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileClose`.
3. Call `FileReadLine` — expect EndOfFile.

---

### Skip after close

Setup: Create a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileClose`.
3. Call `FileSkipLines` with count `1` — expect no error (call does nothing).
