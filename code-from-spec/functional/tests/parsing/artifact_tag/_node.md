---
depends_on:
  - SPEC/functional/logic/parsing/artifact_tag(interface)
output: code-from-spec/functional/tests/parsing/artifact_tag/output.md
---

# SPEC/functional/tests/parsing/artifact_tag

Test cases for the artifact tag component.

# Public

## Test cases

### Happy path

#### Extracts tag from slash-slash comment

Create a file containing:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```
Call `ArtifactTagExtract`. Expect logical name =
`"ROOT/golang/implementation/internal/foo/code(bar)"`,
hash = `"abcdefghijklmnopqrstuvwxyza"`.

#### Extracts tag from hash comment

Create a file containing:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```
Call `ArtifactTagExtract`. Expect logical name =
`"ROOT/some/node(id)"`,
hash = `"123456789012345678901234567"`.

#### Extracts tag from HTML comment

Create a file containing:
```
<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->
```
Call `ArtifactTagExtract`. Expect logical name =
`"ROOT/docs/readme"`,
hash = `"abcdefghijklmnopqrstuvwxyza"`.

#### Stops reading at first match

Create a file with multiple `code-from-spec:` lines.
Call `ArtifactTagExtract`. Expect only the first match
is returned.

#### Tag on non-first line

Create a file where the tag appears on line 3. Call
`ArtifactTagExtract`. Expect the tag is still found.

#### Extra whitespace before logical name

Create a file containing:
```
// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza
```
Call `ArtifactTagExtract`. Expect logical name =
`"ROOT/x(y)"` (leading whitespace trimmed),
hash = `"abcdefghijklmnopqrstuvwxyza"`.

### Edge cases

#### Empty file

Create an empty file. Call `ArtifactTagExtract`.
Expect error NoTagFound.

### Failure cases

#### File does not exist

Call `ArtifactTagExtract` with a non-existent path.
Expect error propagated from file component
(file.FileUnreadable).

#### Propagates path errors

Call `ArtifactTagExtract` with an invalid `PathCfs`
(e.g., `"../../outside"`). Expect error
DirectoryTraversal (propagated from pathutils via
FileOpen).

#### No tag in file

Create a file with no `code-from-spec:` substring.
Call `ArtifactTagExtract`. Expect error NoTagFound.

#### Malformed tag -- no @ separator

Create a file containing:
```
// code-from-spec: ROOT/foo/bar
```
Call `ArtifactTagExtract`. Expect error MalformedTag.

#### Malformed tag -- empty logical name

Create a file containing:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```
Call `ArtifactTagExtract`. Expect error MalformedTag.

#### Malformed tag -- wrong hash length

Create a file containing:
```
// code-from-spec: ROOT/foo(bar)@short
```
Call `ArtifactTagExtract`. Expect error MalformedTag.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `ArtifactTagExtract`.
