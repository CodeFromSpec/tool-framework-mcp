<!-- code-from-spec: ROOT/functional/tests/os/list_files@QMYGK1hm6rlcwN_xZ71fqwFTlhk -->

# Test Specification: ListFiles

## Happy path

### Lists files in a flat directory

Setup: Create a directory containing three files: `a.txt`, `b.txt`, `c.txt`.

Actions:
1. Call `ListFiles` with the directory path.
2. Expect a list of three `PathCfs` values in alphabetical order: `a.txt`, `b.txt`, `c.txt` (paths relative to the project root).

---

### Lists files recursively

Setup: Create the following directory structure:
```
dir/
  alpha.txt
  sub/
    beta.txt
    deep/
      gamma.txt
```

Actions:
1. Call `ListFiles` with the `dir` path.
2. Expect three `PathCfs` values in alphabetical order:
   - `dir/alpha.txt`
   - `dir/sub/beta.txt`
   - `dir/sub/deep/gamma.txt`

---

### Results are sorted alphabetically

Setup: Create a directory containing files `z.txt`, `a.txt`, `m.txt` (created in that order).

Actions:
1. Call `ListFiles` with the directory path.
2. Expect the results in order: `a.txt`, `m.txt`, `z.txt`.

---

## Edge cases

### Empty directory

Setup: Create an empty directory.

Actions:
1. Call `ListFiles` with the empty directory path.
2. Expect an empty list — no error.

---

### Directory with only subdirectories

Setup: Create a directory containing only subdirectories, with no files at any level.

Actions:
1. Call `ListFiles` with the directory path.
2. Expect an empty list — no error.

---

## Failure cases

### Directory does not exist

Setup: No directory is created at the target path.

Actions:
1. Call `ListFiles` with a path to a non-existent directory.
2. Expect error DirectoryNotFound.

---

### Propagates validation errors from PathCfsToOs

Setup: No file or directory is created.

Actions:
1. Call `ListFiles` with an invalid `PathCfs` such as `"../../outside"`.
2. Expect error DirectoryTraversal (propagated from PathUtils).

---

### Propagates conversion errors from PathOsToCfs

Skip this test on platforms where symlinks are not supported.

Setup: Create a directory containing a regular file and a symlink that points to a file outside the project root.

Actions:
1. Call `ListFiles` with the directory path.
2. Expect error ResolvesOutsideRoot (propagated from PathUtils).

---

### Walk error

Skip this test on platforms where directory permissions cannot prevent traversal.

Setup: Create a directory containing a subdirectory with permissions that prevent reading its contents.

Actions:
1. Call `ListFiles` on the parent directory.
2. Expect error WalkError.
