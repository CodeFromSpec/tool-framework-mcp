<!-- code-from-spec: SPEC/functional/tests/parsing/artifact_tag@-YXi8Vv0QSJcYBkCgbR1wsY6PRM -->

# Test Specification: ArtifactTagExtract

---

## Happy Path

---

### TC-01: Extracts tag from slash-slash comment

**Setup**
Create a file containing exactly:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns an ArtifactTag record with:
- logical_name = `"ROOT/golang/implementation/internal/foo/code(bar)"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-02: Extracts tag from hash comment

**Setup**
Create a file containing exactly:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns an ArtifactTag record with:
- logical_name = `"ROOT/some/node(id)"`
- hash = `"123456789012345678901234567"`

---

### TC-03: Extracts tag from HTML comment

**Setup**
Create a file containing exactly:
```
<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns an ArtifactTag record with:
- logical_name = `"ROOT/docs/readme"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-04: Stops reading at first match

**Setup**
Create a file containing multiple `code-from-spec:` lines, for example:
```
// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza
// code-from-spec: ROOT/second/node@abcdefghijklmnopqrstuvwxyzb
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns an ArtifactTag record matching only the first line:
- logical_name = `"ROOT/first/node"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-05: Tag on non-first line

**Setup**
Create a file where the tag appears on line 3, for example:
```
line one
line two
// code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns an ArtifactTag record with:
- logical_name = `"ROOT/docs/readme"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-06: Extra whitespace before logical name

**Setup**
Create a file containing exactly:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns an ArtifactTag record with leading whitespace trimmed:
- logical_name = `"ROOT/x(y)"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

## Edge Cases

---

### TC-07: Empty file

**Setup**
Create an empty file (zero bytes, no content).

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns error `NoTagFound`.

---

## Failure Cases

---

### TC-08: File does not exist

**Setup**
No file is created.

**Action**
Call `ArtifactTagExtract` with a path that does not exist on disk.

**Expected outcome**
Returns error `FileUnreadable`.

---

### TC-09: Propagates path errors

**Setup**
No file is created.

**Action**
Call `ArtifactTagExtract` with an invalid `PathCfs` value such as `"../../outside"`.

**Expected outcome**
Returns error `DirectoryTraversal`, propagated from FileReader/PathUtils via FileOpen.

---

### TC-10: No tag in file

**Setup**
Create a file with content that contains no `code-from-spec:` substring, for example:
```
This file has no artifact tag.
Just regular content.
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns error `NoTagFound`.

---

### TC-11: Malformed tag — no @ separator

**Setup**
Create a file containing exactly:
```
// code-from-spec: ROOT/foo/bar
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns error `MalformedTag`.

---

### TC-12: Malformed tag — empty logical name

**Setup**
Create a file containing exactly:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns error `MalformedTag`.

---

### TC-13: Malformed tag — wrong hash length

**Setup**
Create a file containing exactly:
```
// code-from-spec: ROOT/foo(bar)@short
```

**Action**
Call `ArtifactTagExtract` with the path to that file.

**Expected outcome**
Returns error `MalformedTag`.
