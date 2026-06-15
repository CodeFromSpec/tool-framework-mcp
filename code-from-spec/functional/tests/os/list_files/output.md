<!-- code-from-spec: SPEC/functional/tests/os/list_files@3SGXFz-ZPYaN8ehjGlWqeLERrLQ -->

## Test Suite: ListFiles

---

### TC-01: Lists files in a flat directory

Setup:
- Create a temporary directory under the project root.
- Place three files in it: `a.txt`, `b.txt`, `c.txt`.

Actions:
- Call `ListFiles` with the `PathCfs` of that directory.

Expected outcome:
- Returns a list of three `pathutils.PathCfs` values.
- The values correspond to `a.txt`, `b.txt`, `c.txt` relative to the project root.
- The list is in alphabetical order: `a.txt`, `b.txt`, `c.txt`.

---

### TC-02: Lists files recursively

Setup:
- Create the following directory structure under the project root:
  ```
  dir/
    alpha.txt
    sub/
      beta.txt
      deep/
        gamma.txt
  ```

Actions:
- Call `ListFiles` with the `PathCfs` of `dir`.

Expected outcome:
- Returns a list of three `pathutils.PathCfs` values in alphabetical order:
  `dir/alpha.txt`, `dir/sub/beta.txt`, `dir/sub/deep/gamma.txt`.

---

### TC-03: Results are sorted alphabetically

Setup:
- Create a temporary directory under the project root.
- Place three files in it: `z.txt`, `a.txt`, `m.txt` (created in that order).

Actions:
- Call `ListFiles` with the `PathCfs` of that directory.

Expected outcome:
- Returns a list of three `pathutils.PathCfs` values in order: `a.txt`, `m.txt`, `z.txt`.

---

### TC-04: Empty directory

Setup:
- Create a temporary empty directory under the project root.

Actions:
- Call `ListFiles` with the `PathCfs` of that directory.

Expected outcome:
- Returns an empty list.
- No error is raised.

---

### TC-05: Directory with only subdirectories

Setup:
- Create a temporary directory under the project root.
- Create one or more subdirectories inside it, but no files at any level.

Actions:
- Call `ListFiles` with the `PathCfs` of the top-level directory.

Expected outcome:
- Returns an empty list.
- No error is raised.

---

### TC-06: Directory does not exist

Setup:
- No directory is created; use a path that does not exist on the filesystem.

Actions:
- Call `ListFiles` with a `PathCfs` pointing to that non-existent directory.

Expected outcome:
- Raises error `DirectoryNotFound`.

---

### TC-07: Propagates validation errors from PathCfsToOs

Setup:
- No filesystem setup required.

Actions:
- Call `ListFiles` with an invalid `PathCfs`, for example `"../../outside"`.

Expected outcome:
- Raises error `DirectoryTraversal` (propagated from PathUtils).

---

### TC-08: Propagates conversion errors from PathOsToCfs

Setup:
- Skip this test on platforms where symlinks are not supported.
- Create a temporary directory under the project root.
- Place one regular file in it.
- Create a symlink inside the directory that points to a file outside the project root.

Actions:
- Call `ListFiles` with the `PathCfs` of that directory.

Expected outcome:
- Raises error `ResolvesOutsideRoot` (propagated from PathUtils).

---

### TC-09: Walk error

Setup:
- Skip this test on platforms where directory permissions cannot prevent traversal.
- Create a temporary directory under the project root.
- Create a subdirectory inside it.
- Set the permissions on the subdirectory to prevent reading.

Actions:
- Call `ListFiles` with the `PathCfs` of the parent directory.

Expected outcome:
- Raises error `WalkError`.
