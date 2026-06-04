<!-- code-from-spec: ROOT/functional/tests/parsing/artifact_tag@AX-BgwUs8W6hak5YdCE4oVqG8sI -->

## Test cases for ArtifactTagExtract

---

### Happy path

#### Extracts tag from slash-slash comment

Setup: create a file containing exactly:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns an ArtifactTag record.
- `logical_name` = `"ROOT/golang/implementation/internal/foo/code(bar)"`.
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`.

---

#### Extracts tag from hash comment

Setup: create a file containing exactly:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns an ArtifactTag record.
- `logical_name` = `"ROOT/some/node(id)"`.
- `hash` = `"123456789012345678901234567"`.

---

#### Extracts tag from HTML comment

Setup: create a file containing exactly:
```
<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns an ArtifactTag record.
- `logical_name` = `"ROOT/docs/readme"`.
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`.

---

#### Stops reading at first match

Setup: create a file containing multiple `code-from-spec:` lines, e.g.:
```
// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza
// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaa
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns an ArtifactTag record for the first match only.
- `logical_name` = `"ROOT/first/node"`.
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`.

---

#### Tag on non-first line

Setup: create a file where the `code-from-spec:` tag appears on line 3:
```
line one
line two
// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns an ArtifactTag record.
- `logical_name` = `"ROOT/some/node"`.
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`.

---

#### Extra whitespace before logical name

Setup: create a file containing exactly:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns an ArtifactTag record.
- `logical_name` = `"ROOT/x(y)"` (leading whitespace trimmed).
- `hash` = `"abcdefghijklmnopqrstuvwxyza"`.

---

### Edge cases

#### Empty file

Setup: create an empty file (zero bytes).

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns error `NoTagFound`.

---

### Failure cases

#### File does not exist

Setup: no file is created.

Action: call `ArtifactTagExtract` with a path that does not exist on disk.

Expected outcome:
- Returns error `FileUnreadable`.

---

#### Propagates path errors

Setup: no file is created.

Action: call `ArtifactTagExtract` with an invalid `PathCfs` value such as `"../../outside"`.

Expected outcome:
- Returns error `DirectoryTraversal`, propagated from `FileOpen` via `FileReader`/`PathUtils`.

---

#### No tag in file

Setup: create a file with content that contains no `code-from-spec:` substring, e.g.:
```
this file has no artifact tag
just plain text
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns error `NoTagFound`.

---

#### Malformed tag -- no @ separator

Setup: create a file containing exactly:
```
// code-from-spec: ROOT/foo/bar
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns error `MalformedTag`.

---

#### Malformed tag -- empty logical name

Setup: create a file containing exactly:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns error `MalformedTag`.

---

#### Malformed tag -- wrong hash length

Setup: create a file containing exactly:
```
// code-from-spec: ROOT/foo(bar)@short
```

Action: call `ArtifactTagExtract` with the path of that file.

Expected outcome:
- Returns error `MalformedTag`.
