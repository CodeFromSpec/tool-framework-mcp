<!-- code-from-spec: ROOT/functional/tests/os/list_files@Bupv_bXMNS0WVFbAAcaVe57fmEg -->

## Test cases for ListFiles

### Happy path

#### Lists files in a flat directory

Setup:
  Create a temporary directory containing three files: `a.txt`, `b.txt`, `c.txt`.

Actions:
  Call `ListFiles` with the path to that directory.

Expected outcome:
  Returns a list of three `pathutils.PathCfs` values in alphabetical order:
  `<dir>/a.txt`, `<dir>/b.txt`, `<dir>/c.txt` (relative to the project root).
  No error.

---

#### Lists files recursively

Setup:
  Create the following directory structure:
  ```
  dir/
    alpha.txt
    sub/
      beta.txt
      deep/
        gamma.txt
  ```

Actions:
  Call `ListFiles` with the path to `dir`.

Expected outcome:
  Returns a list of three `pathutils.PathCfs` values in alphabetical order:
  `dir/alpha.txt`, `dir/sub/beta.txt`, `dir/sub/deep/gamma.txt`
  (relative to the project root).
  No error.

---

#### Results are sorted alphabetically

Setup:
  Create a temporary directory containing files named `z.txt`, `a.txt`, `m.txt`.

Actions:
  Call `ListFiles` with the path to that directory.

Expected outcome:
  Returns a list of three `pathutils.PathCfs` values in order:
  `<dir>/a.txt`, `<dir>/m.txt`, `<dir>/z.txt`.
  No error.

---

### Edge cases

#### Empty directory

Setup:
  Create a temporary directory containing no files.

Actions:
  Call `ListFiles` with the path to that directory.

Expected outcome:
  Returns an empty list.
  No error.

---

#### Directory with only subdirectories

Setup:
  Create a directory containing only subdirectories (no files at any level).

Actions:
  Call `ListFiles` with the path to that directory.

Expected outcome:
  Returns an empty list.
  No error.

---

### Failure cases

#### Directory does not exist

Setup:
  No directory is created. Use a path that does not exist.

Actions:
  Call `ListFiles` with the non-existent path.

Expected outcome:
  Returns error `DirectoryNotFound`.

---

#### Propagates validation errors from PathCfsToOs

Setup:
  No directory setup needed. Use an invalid `pathutils.PathCfs` value
  such as `"../../outside"` that resolves outside the project root.

Actions:
  Call `ListFiles` with the invalid path.

Expected outcome:
  Returns error `DirectoryTraversal` propagated from `PathUtils`.

---

#### Propagates conversion errors from PathOsToCfs

Setup:
  Create a directory containing a regular file and a symlink that
  points to a file outside the project root.
  Skip this test case on platforms where symlinks are not supported.

Actions:
  Call `ListFiles` with that directory.

Expected outcome:
  Returns error `ResolvesOutsideRoot` propagated from `PathUtils`.

---

#### Walk error

Setup:
  Create a directory containing a subdirectory with permissions set
  to prevent reading its contents.
  Skip this test case on platforms where directory permissions cannot
  prevent traversal.

Actions:
  Call `ListFiles` on the parent directory.

Expected outcome:
  Returns error `WalkError`.
