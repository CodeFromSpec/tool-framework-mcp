---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/implementation/oslayer(interface)
output: internal/oslayerlisttest/oslayer_list_test.go
---

# SPEC/golang/test/cases/oslayer/list

# Agent

## Test cases

### Happy path

#### Lists files in a flat directory

Setup:
- Create a temporary directory under the project root
  with three files: `a.txt`, `b.txt`, `c.txt`.

Actions:
1. Call `ListAllFiles` with the directory path.

Expected:
- Three CfsPath values in alphabetical order.

#### Lists files recursively

Setup:
- Create directory structure:
  `dir/alpha.txt`, `dir/sub/beta.txt`,
  `dir/sub/deep/gamma.txt`.

Actions:
1. Call `ListAllFiles` with `dir`.

Expected:
- Three CfsPath values in alphabetical order:
  `dir/alpha.txt`, `dir/sub/beta.txt`,
  `dir/sub/deep/gamma.txt`.

#### Results are sorted alphabetically

Setup:
- Create directory with files `z.txt`, `a.txt`,
  `m.txt`.

Actions:
1. Call `ListAllFiles`.

Expected: Order: `a.txt`, `m.txt`, `z.txt`.

### Edge cases

#### Empty directory

Setup:
- Create an empty directory.

Actions:
1. Call `ListAllFiles`.

Expected: Empty list, no error.

#### Hidden files are included

Setup:
- Create directory with `.hidden` and `visible.txt`.

Actions:
1. Call `ListAllFiles`.

Expected:
- Both files returned: `.hidden`, `visible.txt`.

#### Symlink to file within root

Setup:
- Create a file `real.txt` inside the directory.
- Create a symlink `link.txt` pointing to `real.txt`.

Actions:
1. Call `ListAllFiles`.

Expected:
- Both `link.txt` and `real.txt` returned.
Skip on platforms where symlinks are not supported.

#### Directory with only subdirectories

Setup:
- Create directory with only subdirectories (no files
  at any level).

Actions:
1. Call `ListAllFiles`.

Expected: Empty list.

### Failure cases

#### Directory does not exist

Actions:
1. Call `ListAllFiles` with a non-existent path.

Expected: Error `ErrDirectoryNotFound`.

#### Propagates validation errors from CfsPathToOs

Actions:
1. Call `ListAllFiles` with invalid CfsPath
   (e.g., `"../../outside"`).

Expected: Error `ErrDirectoryTraversal`.

#### Propagates conversion errors from OsPathToCfs

Setup:
- Create directory with a regular file and a symlink
  pointing outside the project root.

Actions:
1. Call `ListAllFiles`.

Expected: Error `ErrResolvesOutsideRoot`.
Skip on platforms where symlinks are not supported.

#### Walk error

Setup:
- Create directory with a subdirectory that has
  permissions preventing reading.

Actions:
1. Call `ListAllFiles` on the parent.

Expected: Error `ErrWalkError`. Skip on platforms
where directory permissions cannot prevent traversal.

## Go-specific guidance

- The package name is `oslayerlisttest` (external test
  package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
