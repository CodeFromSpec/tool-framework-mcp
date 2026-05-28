---
depends_on:
  - ROOT/functional/logic/os/file_writer(interface)
outputs:
  - id: file_writer_tests
    path: code-from-spec/functional/tests/os/file_writer/output.md
---

# ROOT/functional/tests/os/file_writer

Test cases for the file writer component.

# Public

## Test cases

### Happy path

#### Writes content to a new file

Call `FileWrite` with a path to a file that does not
exist and content `"hello world"`. Expect the file to
be created with exactly that content.

#### Overwrites an existing file

Create a file with content `"old"`. Call `FileWrite`
with the same path and content `"new"`. Expect the
file to contain `"new"` — the old content is replaced
entirely.

#### Creates intermediate directories

Call `FileWrite` with a path whose parent directories
do not exist (e.g., `"a/b/c/file.txt"`). Expect the
file and all intermediate directories to be created.

#### Preserves UTF-8 content

Call `FileWrite` with content containing non-ASCII
characters: `"café 日本語 🎉"`. Read the file back.
Expect the content to match byte-for-byte.

#### Preserves line endings as received

Call `FileWrite` with content containing CRLF line
endings: `"alpha\r\nbeta\r\n"`. Read the file back.
Expect the content to contain CRLF — no normalization.

#### Writes empty content

Call `FileWrite` with an empty string as content.
Expect a file to be created with zero bytes.

### Failure cases

#### Propagates validation errors from PathCfsToOs

Call `FileWrite` with an invalid `PathCfs` (e.g.,
`"../../outside"`). Expect error "directory traversal"
propagated from `PathCfsToOs`. Expect no file or
directory to be created.

#### Cannot create directory

Call `FileWrite` with a path where an intermediate
directory cannot be created (e.g., a path component
conflicts with an existing file). Expect error
"cannot create directory".

#### Cannot write file

Call `FileWrite` with a path pointing to a directory
that exists (not a file). Expect error
"cannot write file".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `FileWrite`.
