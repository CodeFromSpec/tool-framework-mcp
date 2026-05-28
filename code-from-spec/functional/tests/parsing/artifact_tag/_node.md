---
depends_on:
  - ROOT/functional/logic/parsing/artifact_tag(interface)
outputs:
  - id: artifact_tag_tests
    path: code-from-spec/functional/tests/parsing/artifact_tag/output.md
---

# ROOT/functional/tests/parsing/artifact_tag

Test cases for the artifact tag component.

# Public

## Test cases

### Happy path

#### Extracts tag from slash-slash comment

Create a file containing:
```
// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```
Call ExtractArtifactTag. Expect logical name =
`"ROOT/golang/implementation/internal/foo/code(bar)"`,
hash = `"abcdefghijklmnopqrstuvwxyza"`.

#### Extracts tag from hash comment

Create a file containing:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```
Call ExtractArtifactTag. Expect correct logical name and
hash.

#### Stops reading at first match

Create a file with multiple `code-from-spec:` lines.
Call ExtractArtifactTag. Expect only the first match is
returned.

#### Tag on non-first line

Create a file where the tag appears on line 3. Call
ExtractArtifactTag. Expect the tag is still found.

### Failure cases

#### File does not exist

Call ExtractArtifactTag with a non-existent path. Expect
"file unreadable".

#### No tag in file

Create a file with no `code-from-spec:` substring. Call
ExtractArtifactTag. Expect "no tag found".

#### Malformed tag -- no @ separator

Create a file containing:
```
// code-from-spec: ROOT/foo/bar
```
Call ExtractArtifactTag. Expect "malformed tag".

#### Malformed tag -- empty logical name

Create a file containing:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```
Call ExtractArtifactTag. Expect "malformed tag".

#### Malformed tag -- wrong hash length

Create a file containing:
```
// code-from-spec: ROOT/foo(bar)@short
```
Call ExtractArtifactTag. Expect "malformed tag".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
