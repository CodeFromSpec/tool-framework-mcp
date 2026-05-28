---
depends_on:
  - ROOT/functional/logic/os/list_files(interface)
outputs:
  - id: list_files_tests
    path: code-from-spec/functional/tests/os/list_files/output.md
---

# ROOT/functional/tests/os/list_files

Test cases for the directory listing component.

# Public

## Test cases

### Happy path

#### Lists files in a flat directory

Create a directory containing three files: `a.txt`,
`b.txt`, `c.txt`. Call `ListFiles` with the directory
path. Expect a list of three `PathCfs` values in
alphabetical order: `a.txt`, `b.txt`, `c.txt` (relative
to the project root).

#### Lists files recursively

Create a directory structure:
```
dir/
  alpha.txt
  sub/
    beta.txt
    deep/
      gamma.txt
```
Call `ListFiles` with the `dir` path. Expect three
`PathCfs` values in alphabetical order:
`dir/alpha.txt`, `dir/sub/beta.txt`,
`dir/sub/deep/gamma.txt` (relative to the project root).

#### Results are sorted alphabetically

Create a directory containing files `z.txt`, `a.txt`,
`m.txt`. Call `ListFiles`. Expect the results in order:
`a.txt`, `m.txt`, `z.txt`.

### Edge cases

#### Empty directory

Create an empty directory. Call `ListFiles`. Expect an
empty list — no error.

#### Directory with only subdirectories

Create a directory containing only subdirectories (no
files at any level). Call `ListFiles`. Expect an empty
list.

### Failure cases

#### Directory does not exist

Call `ListFiles` with a path to a non-existent directory.
Expect error "directory not found".

#### Propagates validation errors from PathCfsToOs

Call `ListFiles` with an invalid `PathCfs` (e.g.,
`"../../outside"`). Expect error "directory traversal"
propagated from `PathCfsToOs`.

#### Propagates conversion errors from PathOsToCfs

Create a directory containing a regular file and a
symlink that points to a file outside the project root.
Call `ListFiles`. Expect error "resolves outside root"
propagated from `PathOsToCfs`. Skip on platforms where
symlinks are not supported.

#### Walk error

Create a directory containing a subdirectory with
permissions that prevent reading. Call `ListFiles` on
the parent. Expect error "walk error". Skip on
platforms where directory permissions cannot prevent
traversal.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `ListFiles`.
