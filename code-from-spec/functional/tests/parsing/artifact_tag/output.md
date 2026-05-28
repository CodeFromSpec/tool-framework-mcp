<!-- code-from-spec: ROOT/functional/tests/parsing/artifact_tag@-Ja16QpZ8OA7ZInWKh6l9v230es -->

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
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Returns an `ArtifactTag` record with:
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
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Returns an `ArtifactTag` record with:
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
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Returns an `ArtifactTag` record with:
- `logical_name` = `"ROOT/docs/readme"`
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

### Test: Stops reading at first match

**Setup:**
Create a file containing multiple lines with `code-from-spec:` substrings, for example:
```
// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza
// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaa
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Returns an `ArtifactTag` record matching only the first occurrence:
- `logical_name` = `"ROOT/first/node"`
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

### Test: Tag on non-first line

**Setup:**
Create a file where the tag appears on line 3:
```
line one content
line two content
// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Returns an `ArtifactTag` record with:
- `logical_name` = `"ROOT/some/node"`
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

### Test: Extra whitespace before logical name

**Setup:**
Create a file containing exactly one line with extra spaces after the colon:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Returns an `ArtifactTag` record with leading whitespace trimmed from the logical name:
- `logical_name` = `"ROOT/x(y)"`
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`

---

## Edge Cases

---

### Test: Empty file

**Setup:**
Create a file with no content (zero bytes).

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Raises error `"no tag found"`.

---

## Failure Cases

---

### Test: File does not exist

**Setup:**
No file is created. Use a path that does not exist on disk.

**Action:**
Call `ArtifactTagExtract` with the non-existent path.

**Expected outcome:**
Raises error `"file unreadable"`.

---

### Test: Propagates path errors

**Setup:**
No file is created. Use an invalid `PathCfs` value that would represent a directory traversal, such as `"../../outside"`.

**Action:**
Call `ArtifactTagExtract` with the invalid path.

**Expected outcome:**
Raises error `"directory traversal"` propagated from `FileOpen`.

---

### Test: No tag in file

**Setup:**
Create a file containing content that does not include the `code-from-spec:` substring, for example:
```
This file has no artifact tag at all.
Just some regular text.
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Raises error `"no tag found"`.

---

### Test: Malformed tag -- no @ separator

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec: ROOT/foo/bar
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Raises error `"malformed tag"`.

---

### Test: Malformed tag -- empty logical name

**Setup:**
Create a file containing exactly one line:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Raises error `"malformed tag"`.

---

### Test: Malformed tag -- wrong hash length

**Setup:**
Create a file containing exactly one line where the hash is shorter than expected:
```
// code-from-spec: ROOT/foo(bar)@short
```

**Action:**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome:**
Raises error `"malformed tag"`.
