<!-- code-from-spec: ROOT/functional/tests/parsing/artifact_tag@ezUW8hLOCtFvGx-pRTem769XZxc -->

# Test Specification: ArtifactTagExtract

---

## Happy Path

---

### TC-01: Extracts tag from slash-slash comment

**Setup**

Create a file containing:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns an ArtifactTag record with:
- logical_name = `"ROOT/golang/implementation/internal/foo/code(bar)"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-02: Extracts tag from hash comment

**Setup**

Create a file containing:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns an ArtifactTag record with:
- logical_name = `"ROOT/some/node(id)"`
- hash = `"123456789012345678901234567"`

---

### TC-03: Extracts tag from HTML comment

**Setup**

Create a file containing:
```
<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns an ArtifactTag record with:
- logical_name = `"ROOT/docs/readme"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-04: Stops reading at first match

**Setup**

Create a file containing multiple `code-from-spec:` lines, for example:
```
// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza
// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaz
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns an ArtifactTag record with:
- logical_name = `"ROOT/first/node"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

Only the first matching line is used; subsequent matches are ignored.

---

### TC-05: Tag on non-first line

**Setup**

Create a file where the tag appears on line 3, for example:
```
line one
line two
// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns an ArtifactTag record with:
- logical_name = `"ROOT/some/node"`
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### TC-06: Extra whitespace before logical name

**Setup**

Create a file containing:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns an ArtifactTag record with:
- logical_name = `"ROOT/x(y)"` (leading whitespace trimmed)
- hash = `"abcdefghijklmnopqrstuvwxyza"`

---

## Edge Cases

---

### TC-07: Empty file

**Setup**

Create an empty file (zero bytes).

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns error `NoTagFound`.

---

## Failure Cases

---

### TC-08: File does not exist

**Setup**

No file is created. Use a path that does not exist on the filesystem.

**Action**

Call `ArtifactTagExtract` with the non-existent path.

**Expected Outcome**

Returns error `FileUnreadable`.

---

### TC-09: Propagates path errors

**Setup**

No file is created. Use an invalid `PathCfs` value that contains a directory
traversal sequence, for example `"../../outside"`.

**Action**

Call `ArtifactTagExtract` with that invalid path.

**Expected Outcome**

Returns error `DirectoryTraversal`, propagated from FileReader/PathUtils via
FileOpen.

---

### TC-10: No tag in file

**Setup**

Create a file that contains text but no `code-from-spec:` substring, for
example:
```
This file has no artifact tag at all.
Just some plain content here.
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns error `NoTagFound`.

---

### TC-11: Malformed tag — no @ separator

**Setup**

Create a file containing:
```
// code-from-spec: ROOT/foo/bar
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns error `MalformedTag`.

---

### TC-12: Malformed tag — empty logical name

**Setup**

Create a file containing:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns error `MalformedTag`.

---

### TC-13: Malformed tag — wrong hash length

**Setup**

Create a file containing:
```
// code-from-spec: ROOT/foo(bar)@short
```

**Action**

Call `ArtifactTagExtract` with the path to that file.

**Expected Outcome**

Returns error `MalformedTag`.
