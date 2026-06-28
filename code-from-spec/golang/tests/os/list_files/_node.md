---
depends_on:
  - SPEC/golang/implementation/os/list_files
  - SPEC/golang/implementation/os/path_utils
output: internal/listfiles/listfiles_test.go
---

# SPEC/golang/tests/os/list_files

# Agent

## Test cases

### Happy path

#### Lists files in a flat directory

Setup:
- Create a temporary directory under the project root
  with three files: `a.txt`, `b.txt`, `c.txt`.

Actions:
1. Call `ListFiles` with the directory path.

Expected:
- Three PathCfs values in alphabetical order.

#### Lists files recursively

Setup:
- Create directory structure:
  `dir/alpha.txt`, `dir/sub/beta.txt`,
  `dir/sub/deep/gamma.txt`.

Actions:
1. Call `ListFiles` with `dir`.

Expected:
- Three PathCfs values in alphabetical order:
  `dir/alpha.txt`, `dir/sub/beta.txt`,
  `dir/sub/deep/gamma.txt`.

#### Results are sorted alphabetically

Setup:
- Create directory with files `z.txt`, `a.txt`,
  `m.txt`.

Actions:
1. Call `ListFiles`.

Expected: Order: `a.txt`, `m.txt`, `z.txt`.

### Edge cases

#### Empty directory

Setup:
- Create an empty directory.

Actions:
1. Call `ListFiles`.

Expected: Empty list, no error.

#### Directory with only subdirectories

Setup:
- Create directory with only subdirectories (no files
  at any level).

Actions:
1. Call `ListFiles`.

Expected: Empty list.

### Failure cases

#### Directory does not exist

Actions:
1. Call `ListFiles` with a non-existent path.

Expected: Error `ErrDirectoryNotFound`.

#### Propagates validation errors from PathCfsToOs

Actions:
1. Call `ListFiles` with invalid PathCfs
   (e.g., `"../../outside"`).

Expected: Error `pathutils.ErrDirectoryTraversal`.

#### Propagates conversion errors from PathOsToCfs

Setup:
- Create directory with a regular file and a symlink
  pointing outside the project root.

Actions:
1. Call `ListFiles`.

Expected: Error `pathutils.ErrResolvesOutsideRoot`.
Skip on platforms where symlinks are not supported.

#### Walk error

Setup:
- Create directory with a subdirectory that has
  permissions preventing reading.

Actions:
1. Call `ListFiles` on the parent.

Expected: Error `ErrWalkError`. Skip on platforms
where directory permissions cannot prevent traversal.

## Go-specific guidance

- The package name is `listfiles_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
