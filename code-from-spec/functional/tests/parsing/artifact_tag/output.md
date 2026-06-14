<!-- code-from-spec: ROOT/functional/tests/parsing/artifact_tag@xy-bdV7TeJNnvdOqZbpIa-ae5sI -->

## Test cases for ArtifactTagExtract

---

### Happy path

---

#### TC-1: Extracts tag from slash-slash comment

Setup:
  Create a file containing the single line:
  `// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record with:
  - logical_name = `"ROOT/golang/implementation/internal/foo/code(bar)"`
  - hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-2: Extracts tag from hash comment

Setup:
  Create a file containing the single line:
  `# code-from-spec: ROOT/some/node(id)@123456789012345678901234567`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record with:
  - logical_name = `"ROOT/some/node(id)"`
  - hash = `"123456789012345678901234567"`

---

#### TC-3: Extracts tag from HTML comment

Setup:
  Create a file containing the single line:
  `<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record with:
  - logical_name = `"ROOT/docs/readme"`
  - hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-4: Stops reading at first match

Setup:
  Create a file containing multiple lines each with a `code-from-spec:` tag,
  with different logical names and hashes on each line.

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record corresponding only to the first matching line.
  Subsequent matches are ignored.

---

#### TC-5: Tag on non-first line

Setup:
  Create a file where lines 1 and 2 contain no `code-from-spec:` substring,
  and line 3 contains:
  `// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza`

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record with:
  - logical_name = `"ROOT/some/node"`
  - hash = `"abcdefghijklmnopqrstuvwxyza"`

---

#### TC-6: Extra whitespace before logical name

Setup:
  Create a file containing the single line:
  `// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza`
  (two extra spaces after the colon before the logical name)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Returns an ArtifactTag record with:
  - logical_name = `"ROOT/x(y)"` (leading whitespace trimmed)
  - hash = `"abcdefghijklmnopqrstuvwxyza"`

---

### Edge cases

---

#### TC-7: Empty file

Setup:
  Create an empty file (zero bytes).

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Error NoTagFound is returned.

---

### Failure cases

---

#### TC-8: File does not exist

Setup:
  No file is created. Use a path that does not exist on disk.

Action:
  Call `ArtifactTagExtract` with the non-existent path.

Expected outcome:
  Error FileUnreadable is returned.

---

#### TC-9: Propagates path errors

Setup:
  Prepare an invalid `pathutils.PathCfs` value that represents a path
  attempting directory traversal (e.g., `"../../outside"`).

Action:
  Call `ArtifactTagExtract` with the invalid path.

Expected outcome:
  Error DirectoryTraversal is returned, propagated from FileOpen via PathUtils.

---

#### TC-10: No tag in file

Setup:
  Create a file with content that contains no `code-from-spec:` substring.

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Error NoTagFound is returned.

---

#### TC-11: Malformed tag — no @ separator

Setup:
  Create a file containing the single line:
  `// code-from-spec: ROOT/foo/bar`
  (no `@` character after the logical name)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Error MalformedTag is returned.

---

#### TC-12: Malformed tag — empty logical name

Setup:
  Create a file containing the single line:
  `// code-from-spec: @abcdefghijklmnopqrstuvwxyza`
  (nothing before the `@`)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Error MalformedTag is returned.

---

#### TC-13: Malformed tag — wrong hash length

Setup:
  Create a file containing the single line:
  `// code-from-spec: ROOT/foo(bar)@short`
  (hash segment is shorter than the required 27 characters)

Action:
  Call `ArtifactTagExtract` with the path to that file.

Expected outcome:
  Error MalformedTag is returned.
