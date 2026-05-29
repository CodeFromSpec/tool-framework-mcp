<!-- code-from-spec: ROOT/functional/tests/os/file_reader@RCzuMJy4O8dQJzKcwD2SLXY0FkA -->

# FileReader Test Specification

---

## Happy Path

---

### Test: Opens and reads all lines

**Setup**
Create a file containing three lines with LF endings:
- Line 1: `"alpha"`
- Line 2: `"beta"`
- Line 3: `"gamma"`

**Actions**
1. Call `FileOpen` with the file path. Expect success — receive a FileReader.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect `"gamma"`.
5. Call `FileReadLine`. Expect "end of file".
6. Call `FileClose`.

---

### Test: Normalizes CRLF to LF

**Setup**
Create a file containing two lines with CRLF endings:
- Line 1: `"alpha"` followed by CRLF
- Line 2: `"beta"` followed by CRLF

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"alpha"` — no CR or LF characters in the returned string.
3. Call `FileReadLine`. Expect `"beta"` — no CR or LF characters in the returned string.
4. Call `FileClose`.

---

### Test: Reads file with no trailing newline

**Setup**
Create a file where:
- Line 1: `"alpha"` followed by LF
- Line 2: `"beta"` with no trailing newline

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `"beta"`.
4. Call `FileReadLine`. Expect "end of file".
5. Call `FileClose`.

---

### Test: FileSkipLines advances the reader

**Setup**
Create a file containing five lines with LF endings:
- Line 1: `"one"`
- Line 2: `"two"`
- Line 3: `"three"`
- Line 4: `"four"`
- Line 5: `"five"`

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileSkipLines` with count `2`. Expect no error.
3. Call `FileReadLine`. Expect `"three"`.
4. Call `FileClose`.

---

### Test: FileSkipLines past end of file

**Setup**
Create a file containing two lines with LF endings:
- Line 1: `"one"`
- Line 2: `"two"`

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileSkipLines` with count `10`. Expect no error.
3. Call `FileReadLine`. Expect "end of file".
4. Call `FileClose`.

---

### Test: Preserves leading whitespace

**Setup**
Create a file containing two lines with LF endings:
- Line 1: `"  alpha"` (two leading spaces)
- Line 2: `"    beta"` (four leading spaces)

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"  alpha"` — leading spaces preserved.
3. Call `FileReadLine`. Expect `"    beta"` — leading spaces preserved.
4. Call `FileClose`.

---

### Test: Preserves trailing whitespace

**Setup**
Create a file containing two lines with LF endings:
- Line 1: `"alpha  "` (two trailing spaces)
- Line 2: `"beta   "` (three trailing spaces)

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"alpha  "` — trailing spaces preserved.
3. Call `FileReadLine`. Expect `"beta   "` — trailing spaces preserved.
4. Call `FileClose`.

---

### Test: Preserves internal whitespace

**Setup**
Create a file containing two lines with LF endings:
- Line 1: `"alpha   beta"` (three internal spaces)
- Line 2: `"one\ttwo"` (internal tab character)

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"alpha   beta"` — internal spaces preserved.
3. Call `FileReadLine`. Expect `"one\ttwo"` — internal tab preserved.
4. Call `FileClose`.

---

### Test: Preserves empty lines

**Setup**
Create a file containing four lines with LF endings:
- Line 1: `"alpha"`
- Line 2: `""` (empty)
- Line 3: `""` (empty)
- Line 4: `"beta"`

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"alpha"`.
3. Call `FileReadLine`. Expect `""` — empty string, not skipped.
4. Call `FileReadLine`. Expect `""` — empty string, not skipped.
5. Call `FileReadLine`. Expect `"beta"`.
6. Call `FileClose`.

---

### Test: Preserves non-ASCII characters

**Setup**
Create a UTF-8 encoded file containing three lines with LF endings:
- Line 1: `"café"` (accented character)
- Line 2: `"日本語"` (CJK characters)
- Line 3: `"🎉🚀"` (emoji characters)

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"café"` — accented character preserved.
3. Call `FileReadLine`. Expect `"日本語"` — CJK characters preserved.
4. Call `FileReadLine`. Expect `"🎉🚀"` — emoji characters preserved.
5. Call `FileClose`.

---

## Edge Cases

---

### Test: Empty file

**Setup**
Create an empty file (zero bytes).

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect "end of file" immediately.
3. Call `FileClose`.

---

### Test: Single line without newline

**Setup**
Create a file containing only `"hello"` with no newline character.

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileReadLine`. Expect `"hello"`.
3. Call `FileReadLine`. Expect "end of file".
4. Call `FileClose`.

---

## Failure Cases

---

### Test: File does not exist

**Setup**
No file setup required. Use a path that does not exist on the filesystem.

**Actions**
1. Call `FileOpen` with the non-existent path. Expect "file unreadable".

---

### Test: Read after close

**Setup**
Create a file containing one line with LF ending:
- Line 1: `"alpha"`

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileClose`.
3. Call `FileReadLine`. Expect "end of file".

---

### Test: Skip after close

**Setup**
Create a file containing one line with LF ending:
- Line 1: `"alpha"`

**Actions**
1. Call `FileOpen` with the file path. Expect success.
2. Call `FileClose`.
3. Call `FileSkipLines` with count `1`. Expect no error — the call does nothing.
