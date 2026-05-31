<!-- code-from-spec: ROOT/functional/tests/os/file_reader@Yb-U-rm8CEzeDUMoi6HYUUbhFVA -->

# FileReader Test Specification

## Interface

```
record FileReader
  cfs_path: pathutils.PathCfs

function FileOpen(cfs_path) -> FileReader
  errors:
    - FileUnreadable
    - (PathUtils.*): propagated from PathCfsToOs

function FileReadLine(reader) -> string
  errors:
    - EndOfFile

function FileSkipLines(reader, count)

function FileClose(reader)
```

---

## Happy Path

### TC-01: Opens and reads all lines

**Setup:**
  Create a file containing three lines with LF endings:
  - line 1: `"alpha"`
  - line 2: `"beta"`
  - line 3: `"gamma"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileReadLine` → result 3.
  5. Call `FileReadLine` → result 4.
  6. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"alpha"`
  - result 2 is `"beta"`
  - result 3 is `"gamma"`
  - result 4 raises EndOfFile

---

### TC-02: Normalizes CRLF to LF

**Setup:**
  Create a file containing two lines with CRLF endings:
  - line 1: `"alpha\r\n"`
  - line 2: `"beta\r\n"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"alpha"` — no CR or LF characters
  - result 2 is `"beta"` — no CR or LF characters

---

### TC-03: Reads file with no trailing newline

**Setup:**
  Create a file containing:
  - line 1: `"alpha"` followed by LF
  - line 2: `"beta"` with no trailing newline

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileReadLine` → result 3.
  5. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"alpha"`
  - result 2 is `"beta"`
  - result 3 raises EndOfFile

---

### TC-04: FileSkipLines advances the reader

**Setup:**
  Create a file containing five lines with LF endings:
  - `"one"`, `"two"`, `"three"`, `"four"`, `"five"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileSkipLines` with count 2.
  3. Call `FileReadLine` → result 1.
  4. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"three"`

---

### TC-05: FileSkipLines past end of file

**Setup:**
  Create a file containing two lines with LF endings:
  - `"one"`, `"two"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileSkipLines` with count 10 — expect no error.
  3. Call `FileReadLine` → result 1.
  4. Call `FileClose`.

**Expected outcomes:**
  - `FileSkipLines` completes without raising an error
  - result 1 raises EndOfFile

---

### TC-06: Preserves leading whitespace

**Setup:**
  Create a file containing two lines with LF endings:
  - `"  alpha"` (two leading spaces)
  - `"    beta"` (four leading spaces)

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"  alpha"` — leading spaces preserved
  - result 2 is `"    beta"` — leading spaces preserved

---

### TC-07: Preserves trailing whitespace

**Setup:**
  Create a file containing two lines with LF endings:
  - `"alpha  "` (two trailing spaces)
  - `"beta   "` (three trailing spaces)

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"alpha  "` — trailing spaces preserved
  - result 2 is `"beta   "` — trailing spaces preserved

---

### TC-08: Preserves internal whitespace

**Setup:**
  Create a file containing two lines with LF endings:
  - `"alpha   beta"` (internal spaces)
  - `"one\ttwo"` (internal tab)

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"alpha   beta"` — internal spaces preserved
  - result 2 is `"one\ttwo"` — internal tab preserved

---

### TC-09: Preserves empty lines

**Setup:**
  Create a file containing four lines with LF endings:
  - `"alpha"`, `""`, `""`, `"beta"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileReadLine` → result 3.
  5. Call `FileReadLine` → result 4.
  6. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"alpha"`
  - result 2 is `""` — empty line returned as empty string, not skipped
  - result 3 is `""` — empty line returned as empty string, not skipped
  - result 4 is `"beta"`

---

### TC-10: Preserves non-ASCII characters

**Setup:**
  Create a file containing three lines with LF endings (UTF-8 encoded):
  - `"café"` (accented character)
  - `"日本語"` (CJK characters)
  - `"🎉🚀"` (emoji)

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileReadLine` → result 3.
  5. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"café"` — accented characters pass through unchanged
  - result 2 is `"日本語"` — CJK characters pass through unchanged
  - result 3 is `"🎉🚀"` — emoji pass through unchanged

---

## Edge Cases

### TC-11: Empty file

**Setup:**
  Create an empty file (zero bytes).

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileClose`.

**Expected outcomes:**
  - result 1 raises EndOfFile immediately

---

### TC-12: Single line without newline

**Setup:**
  Create a file containing only `"hello"` with no newline character.

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileReadLine` → result 1.
  3. Call `FileReadLine` → result 2.
  4. Call `FileClose`.

**Expected outcomes:**
  - result 1 is `"hello"`
  - result 2 raises EndOfFile

---

## Failure Cases

### TC-13: File does not exist

**Setup:**
  Identify a path that does not correspond to any existing file.

**Actions:**
  1. Call `FileOpen` with the non-existent path → result 1.

**Expected outcomes:**
  - result 1 raises FileUnreadable

---

### TC-14: Read after close

**Setup:**
  Create a file containing one line with LF ending:
  - `"alpha"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileClose`.
  3. Call `FileReadLine` → result 1.

**Expected outcomes:**
  - result 1 raises EndOfFile

---

### TC-15: Skip after close

**Setup:**
  Create a file containing one line with LF ending:
  - `"alpha"`

**Actions:**
  1. Call `FileOpen` with the file path.
  2. Call `FileClose`.
  3. Call `FileSkipLines` with count 1 — expect no error.

**Expected outcomes:**
  - `FileSkipLines` completes without raising an error — the call does nothing
