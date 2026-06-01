<!-- code-from-spec: ROOT/functional/tests/parsing/artifact_tag@asPYdI70fJXepiQ0v5FFpQmlRLU -->

# Test Specification: ArtifactTagExtract

---

## Happy Path

---

### Test: Extracts tag from slash-slash comment

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- No error.
- Returned `ArtifactTag` has:
  - `logical_name` = `"ROOT/golang/implementation/internal/foo/code(bar)"`
  - `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

### Test: Extracts tag from hash comment

**Setup:**
Create a file containing exactly one line:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- No error.
- Returned `ArtifactTag` has:
  - `logical_name` = `"ROOT/some/node(id)"`
  - `hash` = `"123456789012345678901234567"`

---

### Test: Extracts tag from HTML comment

**Setup:**
Create a file containing exactly one line:
```
<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- No error.
- Returned `ArtifactTag` has:
  - `logical_name` = `"ROOT/docs/readme"`
  - `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

### Test: Stops reading at first match

**Setup:**
Create a file containing multiple lines, each with a `code-from-spec:` tag:
```
// code-from-spec: ROOT/first/node@aaaaaaaaaaaaaaaaaaaaaaaaaa1
// code-from-spec: ROOT/second/node@bbbbbbbbbbbbbbbbbbbbbbbbbbb
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- No error.
- Returned `ArtifactTag` reflects only the first match:
  - `logical_name` = `"ROOT/first/node"`
  - `hash` = `"aaaaaaaaaaaaaaaaaaaaaaaaaa1"`

---

### Test: Tag on non-first line

**Setup:**
Create a file where the tag appears on line 3:
```
Some preamble text
Another line of text
// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- No error.
- Returned `ArtifactTag` has:
  - `logical_name` = `"ROOT/some/node"`
  - `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

### Test: Extra whitespace before logical name

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- No error.
- Returned `ArtifactTag` has leading whitespace trimmed:
  - `logical_name` = `"ROOT/x(y)"`
  - `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

## Edge Cases

---

### Test: Empty file

**Setup:**
Create a file that is empty (zero bytes).

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- Error `NoTagFound`.

---

## Failure Cases

---

### Test: File does not exist

**Setup:**
No file is created. Use a path that does not exist on the filesystem.

**Action:**
Call `ArtifactTagExtract` with the non-existent path.

**Expected outcome:**
- Error `FileUnreadable`.

---

### Test: Propagates path errors

**Setup:**
No file is created. Construct an invalid `PathCfs` value that attempts directory traversal, for example `"../../outside"`.

**Action:**
Call `ArtifactTagExtract` with the invalid `PathCfs`.

**Expected outcome:**
- Error `DirectoryTraversal`, propagated from `FileReader`/`PathUtils` via `FileOpen`.

---

### Test: No tag in file

**Setup:**
Create a file whose contents contain no `code-from-spec:` substring. For example:
```
This file has no artifact tag.
Just some regular content here.
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- Error `NoTagFound`.

---

### Test: Malformed tag -- no @ separator

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec: ROOT/foo/bar
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- Error `MalformedTag`.

---

### Test: Malformed tag -- empty logical name

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- Error `MalformedTag`.

---

### Test: Malformed tag -- wrong hash length

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec: ROOT/foo(bar)@short
```

**Action:**
Call `ArtifactTagExtract` with the path to the file.

**Expected outcome:**
- Error `MalformedTag`.
