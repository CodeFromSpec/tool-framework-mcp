<!-- code-from-spec: ROOT/functional/tests/os/file_reader@RCzuMJy4O8dQJzKcwD2SLXY0FkA -->

# FileReader Test Specification

---

## Happy Path

---

### Test: Opens and reads all lines

**Setup:**
Create a file containing three lines with LF endings:
- Line 1: `"alpha"`
- Line 2: `"beta"`
- Line 3: `"gamma"`

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"alpha"`.
3. Call `FileReadLine` on the reader. Expect `"beta"`.
4. Call `FileReadLine` on the reader. Expect `"gamma"`.
5. Call `FileReadLine` on the reader. Expect "end of file" error.

---

### Test: Normalizes CRLF to LF

**Setup:**
Create a file containing two lines with CRLF endings:
- Line 1: `"alpha"` followed by CRLF
- Line 2: `"beta"` followed by CRLF

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"alpha"` — no CR or LF characters in the returned string.
3. Call `FileReadLine` on the reader. Expect `"beta"` — no CR or LF characters in the returned string.

---

### Test: Reads file with no trailing newline

**Setup:**
Create a file containing:
- Line 1: `"alpha"` followed by LF
- Line 2: `"beta"` with no trailing newline

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"alpha"`.
3. Call `FileReadLine` on the reader. Expect `"beta"`.
4. Call `FileReadLine` on the reader. Expect "end of file" error.

---

### Test: FileSkipLines advances the reader

**Setup:**
Create a file containing five lines with LF endings:
- `"one"`, `"two"`, `"three"`, `"four"`, `"five"`

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileSkipLines` with count `2`.
3. Call `FileReadLine` on the reader. Expect `"three"`.

---

### Test: FileSkipLines past end of file

**Setup:**
Create a file containing two lines:
- `"one"`, `"two"`

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileSkipLines` with count `10`. Expect no error.
3. Call `FileReadLine` on the reader. Expect "end of file" error.

---

### Test: Preserves leading whitespace

**Setup:**
Create a file containing two lines:
- Line 1: `"  alpha"` (two leading spaces)
- Line 2: `"    beta"` (four leading spaces)

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"  alpha"` — leading spaces preserved.
3. Call `FileReadLine` on the reader. Expect `"    beta"` — leading spaces preserved.

---

### Test: Preserves trailing whitespace

**Setup:**
Create a file containing two lines:
- Line 1: `"alpha  "` (two trailing spaces)
- Line 2: `"beta   "` (three trailing spaces)

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"alpha  "` — trailing spaces preserved.
3. Call `FileReadLine` on the reader. Expect `"beta   "` — trailing spaces preserved.

---

### Test: Preserves internal whitespace

**Setup:**
Create a file containing two lines:
- Line 1: `"alpha   beta"` (internal spaces)
- Line 2: `"one\ttwo"` (internal tab)

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"alpha   beta"` — internal spaces preserved.
3. Call `FileReadLine` on the reader. Expect `"one\ttwo"` — internal tab preserved.

---

### Test: Preserves empty lines

**Setup:**
Create a file containing four lines:
- Line 1: `"alpha"`
- Line 2: `""` (empty)
- Line 3: `""` (empty)
- Line 4: `"beta"`

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"alpha"`.
3. Call `FileReadLine` on the reader. Expect `""` — empty line returned as empty string, not skipped.
4. Call `FileReadLine` on the reader. Expect `""` — empty line returned as empty string, not skipped.
5. Call `FileReadLine` on the reader. Expect `"beta"`.

---

### Test: Preserves non-ASCII characters

**Setup:**
Create a file containing three lines (UTF-8 encoded):
- Line 1: `"café"`
- Line 2: `"日本語"`
- Line 3: `"🎉🚀"`

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"café"` — accented characters pass through unchanged.
3. Call `FileReadLine` on the reader. Expect `"日本語"` — CJK characters pass through unchanged.
4. Call `FileReadLine` on the reader. Expect `"🎉🚀"` — emoji pass through unchanged.

---

## Edge Cases

---

### Test: Empty file

**Setup:**
Create an empty file (zero bytes).

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect "end of file" error immediately.

---

### Test: Single line without newline

**Setup:**
Create a file containing only `"hello"` with no newline character.

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileReadLine` on the reader. Expect `"hello"`.
3. Call `FileReadLine` on the reader. Expect "end of file" error.

---

## Failure Cases

---

### Test: File does not exist

**Setup:**
No file is created. Use a path that does not exist on the filesystem.

**Actions:**
1. Call `FileOpen` with the non-existent path. Expect "file unreadable" error.

---

### Test: Read after close

**Setup:**
Create a file containing one line: `"alpha"`.

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileClose` on the reader.
3. Call `FileReadLine` on the reader. Expect "end of file" error.

---

### Test: Skip after close

**Setup:**
Create a file containing one line: `"alpha"`.

**Actions:**
1. Call `FileOpen` with the file path. Expect a FileReader.
2. Call `FileClose` on the reader.
3. Call `FileSkipLines` with count `1`. Expect no error — the call does nothing.
