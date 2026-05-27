<!-- code-from-spec: ROOT/functional/tests/os/file_reader@Vcc1xwJt0SwwipCJ5IYJAA1qans -->

# FileReader Test Specification

---

## Happy Path

---

### Test: Opens and reads all lines

**Setup**
Create a file with three lines using LF endings:
- Line 1: `"alpha"`
- Line 2: `"beta"`
- Line 3: `"gamma"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2
4. Call `FileReadLine(reader)` → result 3
5. Call `FileReadLine(reader)` → result 4

**Expected outcome**
- result 1 = `"alpha"`
- result 2 = `"beta"`
- result 3 = `"gamma"`
- result 4 raises "end of file"

---

### Test: Normalizes CRLF to LF

**Setup**
Create a file with two lines using CRLF endings:
- Line 1: `"alpha"`
- Line 2: `"beta"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2

**Expected outcome**
- result 1 = `"alpha"` (no CR or LF characters present)
- result 2 = `"beta"` (no CR or LF characters present)

---

### Test: Reads file with no trailing newline

**Setup**
Create a file where:
- Line 1: `"alpha"` followed by LF
- Line 2: `"beta"` with no trailing newline

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2
4. Call `FileReadLine(reader)` → result 3

**Expected outcome**
- result 1 = `"alpha"`
- result 2 = `"beta"`
- result 3 raises "end of file"

---

### Test: FileSkipLines advances the reader

**Setup**
Create a file with five lines using LF endings:
- Line 1: `"one"`
- Line 2: `"two"`
- Line 3: `"three"`
- Line 4: `"four"`
- Line 5: `"five"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileSkipLines(reader, 2)`
3. Call `FileReadLine(reader)` → result

**Expected outcome**
- result = `"three"`

---

### Test: FileSkipLines past end of file

**Setup**
Create a file with two lines using LF endings:
- Line 1: `"one"`
- Line 2: `"two"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileSkipLines(reader, 10)` → skip result
3. Call `FileReadLine(reader)` → read result

**Expected outcome**
- skip result raises no error
- read result raises "end of file"

---

### Test: Preserves leading whitespace

**Setup**
Create a file with two lines using LF endings:
- Line 1: `"  alpha"` (two leading spaces)
- Line 2: `"    beta"` (four leading spaces)

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2

**Expected outcome**
- result 1 = `"  alpha"`
- result 2 = `"    beta"`

---

### Test: Preserves trailing whitespace

**Setup**
Create a file with two lines using LF endings:
- Line 1: `"alpha  "` (two trailing spaces)
- Line 2: `"beta   "` (three trailing spaces)

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2

**Expected outcome**
- result 1 = `"alpha  "`
- result 2 = `"beta   "`

---

### Test: Preserves internal whitespace

**Setup**
Create a file with two lines using LF endings:
- Line 1: `"alpha   beta"` (three internal spaces)
- Line 2: `"one\ttwo"` (internal tab character)

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2

**Expected outcome**
- result 1 = `"alpha   beta"`
- result 2 = `"one\ttwo"`

---

### Test: Preserves empty lines

**Setup**
Create a file with four lines using LF endings:
- Line 1: `"alpha"`
- Line 2: `""` (empty)
- Line 3: `""` (empty)
- Line 4: `"beta"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2
4. Call `FileReadLine(reader)` → result 3
5. Call `FileReadLine(reader)` → result 4

**Expected outcome**
- result 1 = `"alpha"`
- result 2 = `""`
- result 3 = `""`
- result 4 = `"beta"`

---

### Test: Preserves non-ASCII characters

**Setup**
Create a file with three lines using LF endings, encoded as UTF-8:
- Line 1: `"café"`
- Line 2: `"日本語"`
- Line 3: `"🎉🚀"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2
4. Call `FileReadLine(reader)` → result 3

**Expected outcome**
- result 1 = `"café"`
- result 2 = `"日本語"`
- result 3 = `"🎉🚀"`

---

## Edge Cases

---

### Test: Empty file

**Setup**
Create an empty file (zero bytes).

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result

**Expected outcome**
- result raises "end of file"

---

### Test: Single line without newline

**Setup**
Create a file containing only `"hello"` with no newline character.

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileReadLine(reader)` → result 1
3. Call `FileReadLine(reader)` → result 2

**Expected outcome**
- result 1 = `"hello"`
- result 2 raises "end of file"

---

## Failure Cases

---

### Test: File does not exist

**Setup**
No file is created. Use a path that does not exist on the filesystem.

**Actions**
1. Call `FileOpen` with the non-existent path → result

**Expected outcome**
- result raises "file unreadable"

---

### Test: Read after close

**Setup**
Create a file with one line using LF endings:
- Line 1: `"alpha"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileClose(reader)`
3. Call `FileReadLine(reader)` → result

**Expected outcome**
- result raises "end of file"

---

### Test: Skip after close

**Setup**
Create a file with one line using LF endings:
- Line 1: `"alpha"`

**Actions**
1. Call `FileOpen` with the file path → `reader`
2. Call `FileClose(reader)`
3. Call `FileSkipLines(reader, 1)` → result

**Expected outcome**
- result raises no error (the call does nothing)
