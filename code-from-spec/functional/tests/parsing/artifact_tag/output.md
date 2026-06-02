<!-- code-from-spec: ROOT/functional/tests/parsing/artifact_tag@vnpqbcltzvFlKzCBAcTtCFSjngU -->

# Test Specification: ArtifactTagExtract

## Happy Path

### Extracts tag from slash-slash comment

Setup: Create a file containing exactly:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Returns an `ArtifactTag` with
`logical_name` = `"ROOT/golang/implementation/internal/foo/code(bar)"` and
`hash` = `"abcdefghijklmnopqrstuvwxyza"`. No error.

---

### Extracts tag from hash comment

Setup: Create a file containing exactly:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Returns an `ArtifactTag` with
`logical_name` = `"ROOT/some/node(id)"` and
`hash` = `"123456789012345678901234567"`. No error.

---

### Extracts tag from HTML comment

Setup: Create a file containing exactly:
```
<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Returns an `ArtifactTag` with
`logical_name` = `"ROOT/docs/readme"` and
`hash` = `"abcdefghijklmnopqrstuvwxyza"`. No error.

---

### Stops reading at first match

Setup: Create a file with multiple `code-from-spec:` lines, e.g.:
```
// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza
// code-from-spec: ROOT/second/node@bcdefghijklmnopqrstuvwxyzab
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Returns the `ArtifactTag` from the first matching line only:
`logical_name` = `"ROOT/first/node"`, `hash` = `"abcdefghijklmnopqrstuvwxyza"`. No error.

---

### Tag on non-first line

Setup: Create a file where the tag appears on line 3:
```
line one
line two
// code-from-spec: ROOT/x/y@abcdefghijklmnopqrstuvwxyza
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Returns an `ArtifactTag` with
`logical_name` = `"ROOT/x/y"` and
`hash` = `"abcdefghijklmnopqrstuvwxyza"`. No error.

---

### Extra whitespace before logical name

Setup: Create a file containing exactly:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Returns an `ArtifactTag` with
`logical_name` = `"ROOT/x(y)"` (leading whitespace trimmed) and
`hash` = `"abcdefghijklmnopqrstuvwxyza"`. No error.

---

## Edge Cases

### Empty file

Setup: Create a file with no content.

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Error `NoTagFound`.

---

## Failure Cases

### File does not exist

Setup: None.

Action: Call `ArtifactTagExtract` with a path that does not exist on disk.

Expected outcome: Error `FileUnreadable`.

---

### Propagates path errors

Setup: None.

Action: Call `ArtifactTagExtract` with an invalid `PathCfs` such as `"../../outside"`.

Expected outcome: Error `DirectoryTraversal` propagated from `FileReader`/`PathUtils` via `FileOpen`.

---

### No tag in file

Setup: Create a file with content that contains no `code-from-spec:` substring, e.g.:
```
This is just a regular file.
No artifact tag here.
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Error `NoTagFound`.

---

### Malformed tag — no @ separator

Setup: Create a file containing exactly:
```
// code-from-spec: ROOT/foo/bar
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Error `MalformedTag`.

---

### Malformed tag — empty logical name

Setup: Create a file containing exactly:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Error `MalformedTag`.

---

### Malformed tag — wrong hash length

Setup: Create a file containing exactly:
```
// code-from-spec: ROOT/foo(bar)@short
```

Action: Call `ArtifactTagExtract` with the path to that file.

Expected outcome: Error `MalformedTag`.
