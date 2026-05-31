<!-- code-from-spec: ROOT/functional/tests/mcp_tools/hash_fragment@g9Rca0SAI3ClyhF1lEzZIDyj_As -->

# Test Specification: MCPHashFragment

Fragment hashes use SHA-1 encoded as base64url (RFC 4648 §5, no padding),
always 27 characters. The input to SHA-1 is the extracted lines, each with
`\n` appended (including the last line).

---

## Happy Path

### Test: Hashes a valid line range

**Setup:**
Create a file with exactly 5 lines of known content, for example:
- Line 1: `"alpha"`
- Line 2: `"bravo"`
- Line 3: `"charlie"`
- Line 4: `"delta"`
- Line 5: `"echo"`

**Action:**
Call `MCPHashFragment` with:
- `path` = path to the file above
- `lines` = `"2-4"`

**Expected outcome:**
Success. The result is a string of exactly 27 characters.
The value matches the base64url-encoded (no padding) SHA-1 digest of:
`"bravo\ncharlie\ndelta\n"`

---

### Test: Single line range

**Setup:**
Create a file with exactly 5 lines of known content (same as above).

**Action:**
Call `MCPHashFragment` with:
- `path` = path to the file
- `lines` = `"3-3"`

**Expected outcome:**
Success. The result is a string of exactly 27 characters.
The value matches the base64url-encoded (no padding) SHA-1 digest of:
`"charlie\n"`

---

### Test: First line of file

**Setup:**
Create a file with exactly 5 lines of known content (same as above).

**Action:**
Call `MCPHashFragment` with:
- `path` = path to the file
- `lines` = `"1-1"`

**Expected outcome:**
Success. The result is a string of exactly 27 characters.
The value matches the base64url-encoded (no padding) SHA-1 digest of:
`"alpha\n"`

---

### Test: Last line of file

**Setup:**
Create a file with exactly 5 lines of known content (same as above).

**Action:**
Call `MCPHashFragment` with:
- `path` = path to the file
- `lines` = `"5-5"`

**Expected outcome:**
Success. The result is a string of exactly 27 characters.
The value matches the base64url-encoded (no padding) SHA-1 digest of:
`"echo\n"`

---

### Test: Hash is deterministic

**Setup:**
Create a file with known content.

**Action:**
Call `MCPHashFragment` twice with the same `path` and `lines` values.

**Expected outcome:**
Both calls succeed. The two returned strings are identical.

---

## Error Cases

### Test: File does not exist

**Setup:**
No file is created.

**Action:**
Call `MCPHashFragment` with:
- `path` = `"nonexistent.go"`
- `lines` = `"1-5"`

**Expected outcome:**
Error `FileUnreadable` is returned, propagated from `FileReader` via `FileOpen`.

---

### Test: Invalid line range format — not a range

**Action:**
Call `MCPHashFragment` with:
- `path` = any valid existing file path
- `lines` = `"abc"`

**Expected outcome:**
Error `InvalidLineRange` is returned.

---

### Test: Start greater than end

**Action:**
Call `MCPHashFragment` with:
- `path` = any valid existing file path
- `lines` = `"5-2"`

**Expected outcome:**
Error `InvalidLineRange` is returned.

---

### Test: Start less than 1

**Action:**
Call `MCPHashFragment` with:
- `path` = any valid existing file path
- `lines` = `"0-5"`

**Expected outcome:**
Error `InvalidLineRange` is returned.

---

### Test: Line range out of bounds

**Setup:**
Create a file with exactly 3 lines.

**Action:**
Call `MCPHashFragment` with:
- `path` = path to the file
- `lines` = `"1-10"`

**Expected outcome:**
Error `InvalidLineRange` is returned, because `end` (10) exceeds the file's
line count (3).

---

### Test: Empty path

**Action:**
Call `MCPHashFragment` with:
- `path` = `""`
- `lines` = `"1-5"`

**Expected outcome:**
Error `PathEmpty` is returned, propagated from `PathUtils` via `PathValidateCfs`.

---

### Test: Path traversal

**Action:**
Call `MCPHashFragment` with:
- `path` = `"../../etc/passwd"`
- `lines` = `"1-5"`

**Expected outcome:**
Error `DirectoryTraversal` is returned, propagated from `PathUtils` via
`PathValidateCfs`.
