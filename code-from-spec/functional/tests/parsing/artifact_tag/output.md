<!-- code-from-spec: SPEC/functional/tests/parsing/artifact_tag@Uhrb7YEP9pr7-Z4xmT6L3RqJuBY -->

## Test suite: ArtifactTagExtract

---

### Happy path

---

#### TC-01: Extracts tag from slash-slash comment

Setup:
  Create a file containing exactly:
    `// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record where:
    logical_name = `"ROOT/golang/implementation/internal/foo/code(bar)"`
    hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-02: Extracts tag from hash comment

Setup:
  Create a file containing exactly:
    `# code-from-spec: ROOT/some/node(id)@123456789012345678901234567`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record where:
    logical_name = `"ROOT/some/node(id)"`
    hash = `"123456789012345678901234567"`

---

#### TC-03: Extracts tag from HTML comment

Setup:
  Create a file containing exactly:
    `<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record where:
    logical_name = `"ROOT/docs/readme"`
    hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-04: Stops reading at first match

Setup:
  Create a file containing multiple lines that each include `code-from-spec:`, for example:
    Line 1: `// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza`
    Line 2: `// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaz`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record matching only the first line:
    logical_name = `"ROOT/first/node"`
    hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-05: Tag on non-first line

Setup:
  Create a file where the first two lines contain no `code-from-spec:` substring, and line 3 contains:
    `// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record where:
    logical_name = `"ROOT/some/node"`
    hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-06: Extra whitespace before logical name

Setup:
  Create a file containing exactly:
    `// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza`
  (Note: multiple spaces between `:` and the logical name.)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record where:
    logical_name = `"ROOT/x(y)"` (leading whitespace trimmed)
    hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### Edge cases

---

#### TC-07: Empty file

Setup:
  Create an empty file (zero bytes).

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns error NoTagFound.

---

### Failure cases

---

#### TC-08: File does not exist

Setup:
  No file is created. Use a path that does not exist on the filesystem.

Action:
  Call `ArtifactTagExtract` with that non-existent path.

Expected outcome:
  Returns error FileUnreadable.

---

#### TC-09: Propagates path errors

Setup:
  Construct an invalid PathCfs value that represents a directory traversal, such as `"../../outside"`.

Action:
  Call `ArtifactTagExtract` with that invalid path.

Expected outcome:
  Returns error DirectoryTraversal, propagated from FileReader/PathUtils via FileOpen.

---

#### TC-10: No tag in file

Setup:
  Create a file whose contents contain no `code-from-spec:` substring anywhere.

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns error NoTagFound.

---

#### TC-11: Malformed tag — no @ separator

Setup:
  Create a file containing exactly:
    `// code-from-spec: ROOT/foo/bar`
  (No `@` character in the tag value.)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns error MalformedTag.

---

#### TC-12: Malformed tag — empty logical name

Setup:
  Create a file containing exactly:
    `// code-from-spec: @abcdefghijklmnopqrstuvwxyza`
  (Nothing before the `@`.)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns error MalformedTag.

---

#### TC-13: Malformed tag — wrong hash length

Setup:
  Create a file containing exactly:
    `// code-from-spec: ROOT/foo(bar)@short`
  (The hash `"short"` is fewer than 27 characters.)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns error MalformedTag.
