---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/artifact_tag
output: internal/artifacttag/artifacttag_test.go
---

# SPEC/golang/tests/parsing/artifact_tag

# Agent

## Test cases

### Happy path

#### Extracts tag from slash-slash comment

Setup:
- Create a file containing:
  `// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza`

Actions:
1. Call `ArtifactTagExtract` with the path.

Expected:
- LogicalName = `"ROOT/golang/implementation/internal/foo/code(bar)"`
- Hash = `"abcdefghijklmnopqrstuvwxyza"`

#### Extracts tag from hash comment

Setup:
- Create a file containing:
  `# code-from-spec: ROOT/some/node(id)@123456789012345678901234567`

Actions:
1. Call `ArtifactTagExtract` with the path.

Expected:
- LogicalName = `"ROOT/some/node(id)"`
- Hash = `"123456789012345678901234567"`

#### Extracts tag from HTML comment

Setup:
- Create a file containing:
  `<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->`

Actions:
1. Call `ArtifactTagExtract` with the path.

Expected:
- LogicalName = `"ROOT/docs/readme"`
- Hash = `"abcdefghijklmnopqrstuvwxyza"`

#### Stops reading at first match

Setup:
- Create a file with two `code-from-spec:` lines.

Actions:
1. Call `ArtifactTagExtract`.

Expected: Returns only the first match.

#### Tag on non-first line

Setup:
- Create a file where the tag appears on line 3.

Actions:
1. Call `ArtifactTagExtract`.

Expected: Tag is found.

#### Extra whitespace before logical name

Setup:
- Create a file containing:
  `// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza`

Actions:
1. Call `ArtifactTagExtract`.

Expected:
- LogicalName = `"ROOT/x(y)"` (whitespace trimmed)
- Hash = `"abcdefghijklmnopqrstuvwxyza"`

### Edge cases

#### Empty file

Setup:
- Create an empty file.

Actions:
1. Call `ArtifactTagExtract`.

Expected: Error `ErrNoTagFound`.

### Failure cases

#### File does not exist

Actions:
1. Call `ArtifactTagExtract` with a non-existent path.

Expected: Error `file.ErrFileUnreadable`.

#### Propagates path errors

Actions:
1. Call `ArtifactTagExtract` with an invalid PathCfs
   (e.g., `"../../outside"`).

Expected: Error `pathutils.ErrDirectoryTraversal`.

#### No tag in file

Setup:
- Create a file with no `code-from-spec:` substring.

Actions:
1. Call `ArtifactTagExtract`.

Expected: Error `ErrNoTagFound`.

#### Malformed tag — no @ separator

Setup:
- Create a file containing:
  `// code-from-spec: ROOT/foo/bar`

Actions:
1. Call `ArtifactTagExtract`.

Expected: Error `ErrMalformedTag`.

#### Malformed tag — empty logical name

Setup:
- Create a file containing:
  `// code-from-spec: @abcdefghijklmnopqrstuvwxyza`

Actions:
1. Call `ArtifactTagExtract`.

Expected: Error `ErrMalformedTag`.

#### Malformed tag — wrong hash length

Setup:
- Create a file containing:
  `// code-from-spec: ROOT/foo(bar)@short`

Actions:
1. Call `ArtifactTagExtract`.

Expected: Error `ErrMalformedTag`.

## Go-specific guidance

- The package name is `artifacttag_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
