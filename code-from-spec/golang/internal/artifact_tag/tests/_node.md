---
outputs:
  - id: artifacttag_test
    path: internal/artifacttag/artifacttag_test.go
---

# ROOT/golang/internal/artifact_tag/tests

Test cases for the artifacttag package.

# Agent

## Context

Each test uses `t.TempDir()` to create an isolated temporary
directory. Test files are created with controlled content.
Use table-driven tests where appropriate.

## Happy Path

### Extracts tag from Go comment

Create a file containing:
```
// code-from-spec: ROOT/golang/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza
```
Expect `LogicalName` = `"ROOT/golang/internal/foo/code(bar)"`,
`Hash` = `"abcdefghijklmnopqrstuvwxyza"`.

### Extracts tag from hash comment

Create a file containing:
```
# code-from-spec: ROOT/some/node(id)@123456789012345678901234567
```
Expect correct logical name and hash.

### Stops reading at first match

Create a file with multiple `code-from-spec:` lines.
Expect only the first match is returned.

### Tag on non-first line

Create a file where the tag appears on line 3.
Expect the tag is still found.

## Failure Cases

### File does not exist

Call `ExtractArtifactTag` with a non-existent path.
Expect `errors.Is(err, ErrFileUnreadable)`.

### No tag in file

Create a file with no `code-from-spec:` substring.
Expect `errors.Is(err, ErrNoTagFound)`.

### Malformed tag — no @ separator

Create a file containing:
```
// code-from-spec: ROOT/foo/bar
```
Expect `errors.Is(err, ErrMalformedTag)`.

### Malformed tag — empty logical name

Create a file containing:
```
// code-from-spec: @abcdefghijklmnopqrstuvwxyza
```
Expect `errors.Is(err, ErrMalformedTag)`.

### Malformed tag — wrong hash length

Create a file containing:
```
// code-from-spec: ROOT/foo(bar)@short
```
Expect `errors.Is(err, ErrMalformedTag)`.
