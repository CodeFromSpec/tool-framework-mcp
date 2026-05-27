<!-- code-from-spec: ROOT/functional/tests/utils/file_reader@J6isUPEeJtFpmZnEIjB_qSTCr5U -->

## Test Cases: FileReader

---

### Happy Path

---

#### Opens and reads all lines

**Setup**
Create a file containing multiple lines separated by LF endings, for example:
- Line 1: `"alpha"`
- Line 2: `"beta"`
- Line 3: `"gamma"`

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `ReadLine` repeatedly until the end.

**Expected outcome**
- First call returns `"alpha"`.
- Second call returns `"beta"`.
- Third call returns `"gamma"`.
- Fourth call raises `"end of file"`.

---

#### Normalizes CRLF to LF

**Setup**
Create a file containing lines separated by CRLF endings, for example:
- Line 1: `"alpha"` followed by CRLF
- Line 2: `"beta"` followed by CRLF

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `ReadLine` twice.

**Expected outcome**
- First call returns `"alpha"` — no CR or LF characters present.
- Second call returns `"beta"` — no CR or LF characters present.

---

#### Reads file with no trailing newline

**Setup**
Create a file where the last line has no trailing newline, for example:
- Line 1: `"alpha"` followed by LF
- Line 2: `"beta"` with no newline at end

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `ReadLine` once.
3. Call `ReadLine` again.
4. Call `ReadLine` a third time.

**Expected outcome**
- First call returns `"alpha"`.
- Second call returns `"beta"`.
- Third call raises `"end of file"`.

---

#### SkipLines advances the reader

**Setup**
Create a file containing 5 lines:
- Line 1: `"one"`
- Line 2: `"two"`
- Line 3: `"three"`
- Line 4: `"four"`
- Line 5: `"five"`

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `SkipLines` with count `2`.
3. Call `ReadLine`.

**Expected outcome**
- `ReadLine` returns `"three"`.

---

#### SkipLines past end of file

**Setup**
Create a file containing 2 lines:
- Line 1: `"one"`
- Line 2: `"two"`

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `SkipLines` with count `10`.
3. Call `ReadLine`.

**Expected outcome**
- `SkipLines` completes without error.
- `ReadLine` raises `"end of file"`.

---

### Edge Cases

---

#### Empty file

**Setup**
Create a file with no content (zero bytes).

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `ReadLine`.

**Expected outcome**
- `ReadLine` raises `"end of file"` immediately.

---

#### Single line without newline

**Setup**
Create a file containing exactly `"hello"` with no newline character.

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `ReadLine`.
3. Call `ReadLine` again.

**Expected outcome**
- First call returns `"hello"`.
- Second call raises `"end of file"`.

---

### Failure Cases

---

#### File does not exist

**Setup**
No file is created. Use a path that does not exist on the filesystem.

**Actions**
1. Call `OpenFileReader` with the non-existent path.

**Expected outcome**
- `OpenFileReader` raises `"file unreadable"`.

---

### Close Behavior

---

#### Reading after Close raises end of file

**Setup**
Create a file containing at least one line, for example:
- Line 1: `"alpha"`

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `Close` on the reader.
3. Call `ReadLine`.

**Expected outcome**
- `ReadLine` raises `"end of file"`.

---

#### SkipLines after Close raises end of file

**Setup**
Create a file containing at least one line, for example:
- Line 1: `"alpha"`

**Actions**
1. Call `OpenFileReader` with the file path.
2. Call `Close` on the reader.
3. Call `SkipLines` with count `1`.

**Expected outcome**
- `SkipLines` raises `"end of file"`.
