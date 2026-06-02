<!-- code-from-spec: ROOT/functional/tests/os/file_reader@nZJS9vADfXZCX234xZLWCCsxpY8 -->

# Test Specification: FileReader

## Happy path

### Opens and reads all lines

Setup: Create a file containing three lines: `"alpha"`, `"beta"`, `"gamma"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"alpha"`.
3. Call `FileReadLine` → expect `"beta"`.
4. Call `FileReadLine` → expect `"gamma"`.
5. Call `FileReadLine` → expect error EndOfFile.

---

### Normalizes CRLF to LF

Setup: Create a file containing `"alpha"` and `"beta"` with CRLF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"alpha"` (no CR or LF characters in the returned string).
3. Call `FileReadLine` → expect `"beta"` (no CR or LF characters in the returned string).

---

### Reads file with no trailing newline

Setup: Create a file containing `"alpha"` (with LF) and `"beta"` (no trailing newline).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"alpha"`.
3. Call `FileReadLine` → expect `"beta"`.
4. Call `FileReadLine` → expect error EndOfFile.

---

### FileSkipLines advances the reader

Setup: Create a file containing `"one"`, `"two"`, `"three"`, `"four"`, `"five"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileSkipLines` with count `2`.
3. Call `FileReadLine` → expect `"three"`.

---

### FileSkipLines past end of file

Setup: Create a file containing `"one"`, `"two"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileSkipLines` with count `10` → expect no error.
3. Call `FileReadLine` → expect error EndOfFile.

---

### Preserves leading whitespace

Setup: Create a file containing `"  alpha"` and `"    beta"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"  alpha"` (leading spaces preserved).
3. Call `FileReadLine` → expect `"    beta"` (leading spaces preserved).

---

### Preserves trailing whitespace

Setup: Create a file containing `"alpha  "` and `"beta   "` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"alpha  "` (trailing spaces preserved).
3. Call `FileReadLine` → expect `"beta   "` (trailing spaces preserved).

---

### Preserves internal whitespace

Setup: Create a file containing `"alpha   beta"` and `"one\ttwo"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"alpha   beta"` (internal spaces preserved).
3. Call `FileReadLine` → expect `"one\ttwo"` (tab preserved).

---

### Preserves empty lines

Setup: Create a file containing `"alpha"`, `""`, `""`, `"beta"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"alpha"`.
3. Call `FileReadLine` → expect `""`.
4. Call `FileReadLine` → expect `""`.
5. Call `FileReadLine` → expect `"beta"`.

---

### Preserves non-ASCII characters

Setup: Create a file containing `"café"`, `"日本語"`, `"🎉🚀"` with LF endings, encoded as UTF-8.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"café"`.
3. Call `FileReadLine` → expect `"日本語"`.
4. Call `FileReadLine` → expect `"🎉🚀"`.

---

## Edge cases

### Empty file

Setup: Create an empty file.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect error EndOfFile immediately.

---

### Single line without newline

Setup: Create a file containing only `"hello"` with no newline character.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` → expect `"hello"`.
3. Call `FileReadLine` → expect error EndOfFile.

---

## Failure cases

### File does not exist

Setup: No file is created.

Actions:
1. Call `FileOpen` with a path to a non-existent file → expect error FileUnreadable.

---

### Read after close

Setup: Create a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileClose`.
3. Call `FileReadLine` → expect error EndOfFile.

---

### Skip after close

Setup: Create a file containing `"alpha"`.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileClose`.
3. Call `FileSkipLines` with count `1` → expect no error (the call does nothing).
