<!-- code-from-spec: ROOT/functional/tests/mcp_tools/hash_fragment@y67Tex-O87yklzDh28sxL-UjI1s -->

# Test Specification: MCPHashFragment

## Hashing Convention

The input to SHA-1 is each extracted line with `\n` appended (including the
last line). The digest is encoded as base64url (RFC 4648 §5, no padding),
always producing a 27-character string.

---

## Happy Path

### Test: Hashes a valid line range

Setup:
  Create a file containing exactly 5 lines of known content, e.g.:
    line 1: "alpha"
    line 2: "bravo"
    line 3: "charlie"
    line 4: "delta"
    line 5: "echo"

Action:
  Call MCPHashFragment with path = <path to the file>, lines = "2-4".

Expected outcome:
  1. No error is returned.
  2. The result is a string of exactly 27 characters.
  3. The result equals the base64url-no-padding SHA-1 of the byte sequence:
       "bravo\n" + "charlie\n" + "delta\n"

---

### Test: Single line range

Setup:
  Create a file containing exactly 5 lines of known content.

Action:
  Call MCPHashFragment with path = <path to the file>, lines = "3-3".

Expected outcome:
  1. No error is returned.
  2. The result is a string of exactly 27 characters.
  3. The result equals the base64url-no-padding SHA-1 of the byte sequence:
       <line 3 content> + "\n"

---

### Test: First line of file

Setup:
  Create a file containing exactly 5 lines of known content.

Action:
  Call MCPHashFragment with path = <path to the file>, lines = "1-1".

Expected outcome:
  1. No error is returned.
  2. The result is a string of exactly 27 characters.
  3. The result equals the base64url-no-padding SHA-1 of the byte sequence:
       <line 1 content> + "\n"

---

### Test: Last line of file

Setup:
  Create a file containing exactly 5 lines of known content.

Action:
  Call MCPHashFragment with path = <path to the file>, lines = "5-5".

Expected outcome:
  1. No error is returned.
  2. The result is a string of exactly 27 characters.
  3. The result equals the base64url-no-padding SHA-1 of the byte sequence:
       <line 5 content> + "\n"

---

### Test: Hash is deterministic

Setup:
  Create a file with known content.

Actions:
  1. Call MCPHashFragment with path = <path to the file>, lines = "1-3".
     Save the result as result-1.
  2. Call MCPHashFragment with the same path and lines = "1-3".
     Save the result as result-2.

Expected outcome:
  1. No error is returned for either call.
  2. result-1 equals result-2.

---

## Error Cases

### Test: File does not exist

Setup:
  No file is created.

Action:
  Call MCPHashFragment with path = "nonexistent.go", lines = "1-5".

Expected outcome:
  Error FileUnreadable is returned (propagated from FileReader via FileOpen).

---

### Test: Invalid line range format — not a range

Setup:
  None required.

Action:
  Call MCPHashFragment with path = <any valid file path>, lines = "abc".

Expected outcome:
  Error InvalidLineRange is returned.

---

### Test: Start greater than end

Setup:
  None required.

Action:
  Call MCPHashFragment with path = <any valid file path>, lines = "5-2".

Expected outcome:
  Error InvalidLineRange is returned.

---

### Test: Start less than 1

Setup:
  None required.

Action:
  Call MCPHashFragment with path = <any valid file path>, lines = "0-5".

Expected outcome:
  Error InvalidLineRange is returned.

---

### Test: Line range out of bounds

Setup:
  Create a file containing exactly 3 lines.

Action:
  Call MCPHashFragment with path = <path to the file>, lines = "1-10".

Expected outcome:
  Error InvalidLineRange is returned (end exceeds the file's line count).

---

### Test: Empty path

Setup:
  None required.

Action:
  Call MCPHashFragment with path = "", lines = "1-5".

Expected outcome:
  Error PathEmpty is returned (propagated from PathUtils via PathValidateCfs).

---

### Test: Path traversal

Setup:
  None required.

Action:
  Call MCPHashFragment with path = "../../etc/passwd", lines = "1-5".

Expected outcome:
  Error DirectoryTraversal is returned (propagated from PathUtils via PathValidateCfs).
