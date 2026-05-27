---
depends_on:
  - ROOT/functional/logic/utils/file_reader(interface)
outputs:
  - id: file_reader_tests
    path: code-from-spec/functional/tests/utils/file_reader/output.md
---

# ROOT/functional/tests/utils/file_reader

Test cases for the file reader component.

# Public

## Test cases

### Happy path

#### Opens and reads all lines

Create a file with multiple lines (LF endings). Open it,
then read lines repeatedly. Expect each line returned
without terminator. After the last line, reading raises
"end of file".

#### Normalizes CRLF to LF

Create a file with CRLF line endings. Expect reading
returns lines without any CR or LF characters.

#### Reads file with no trailing newline

Create a file where the last line has no trailing newline.
Expect the last line is returned normally, and the next
read raises "end of file".

#### SkipLines advances the reader

Create a file with 5 lines. Skip 2 lines, then read.
Expect the third line is returned.

#### SkipLines past end of file

Create a file with 2 lines. Skip 10 lines. Expect no
error. Then reading raises "end of file".

### Edge cases

#### Empty file

Create an empty file. Open it, then read. Expect "end of
file" immediately.

#### Single line without newline

Create a file containing just `"hello"` with no newline.
Expect reading returns `"hello"`, then "end of file".

### Failure cases

#### File does not exist

Open a non-existent path. Expect "file unreadable".

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
