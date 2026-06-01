<!-- code-from-spec: ROOT/functional/tests/os/list_files@wHsX2CkyIvIlDrH32-7PkBw6YUU -->

# Test Specification: ListFiles

## Interface

```
function ListFiles(cfs_path: pathutils.PathCfs) -> list of pathutils.PathCfs
  errors:
    - DirectoryNotFound: the directory does not exist.
    - WalkError: a filesystem error occurred while traversing.
    - (PathUtils.*): propagated from pathutils.PathCfsToOs.
    - (PathUtils.*): propagated from PathOsToCfs.
```

---

## Happy Path

### Test: Lists files in a flat directory

**Setup:**
- Create a temporary directory within the project root.
- Inside that directory, create three files: `a.txt`, `b.txt`, `c.txt`.

**Action:**
- Call `ListFiles` with the `PathCfs` for the temporary directory.

**Expected outcome:**
- Returns a list of three `PathCfs` values.
- The list is in alphabetical order: `<dir>/a.txt`, `<dir>/b.txt`, `<dir>/c.txt`
  (paths relative to the project root).
- No error is returned.

---

### Test: Lists files recursively

**Setup:**
- Create the following directory structure within the project root:
  ```
  dir/
    alpha.txt
    sub/
      beta.txt
      deep/
        gamma.txt
  ```

**Action:**
- Call `ListFiles` with the `PathCfs` for `dir`.

**Expected outcome:**
- Returns a list of three `PathCfs` values in alphabetical order:
  `dir/alpha.txt`, `dir/sub/beta.txt`, `dir/sub/deep/gamma.txt`.
- No error is returned.

---

### Test: Results are sorted alphabetically

**Setup:**
- Create a temporary directory within the project root.
- Inside that directory, create three files in non-alphabetical order: `z.txt`, `a.txt`, `m.txt`.

**Action:**
- Call `ListFiles` with the `PathCfs` for the temporary directory.

**Expected outcome:**
- Returns a list of three `PathCfs` values in strictly alphabetical order:
  `<dir>/a.txt`, `<dir>/m.txt`, `<dir>/z.txt`.
- No error is returned.

---

## Edge Cases

### Test: Empty directory

**Setup:**
- Create a temporary directory within the project root that contains no files or subdirectories.

**Action:**
- Call `ListFiles` with the `PathCfs` for the empty directory.

**Expected outcome:**
- Returns an empty list.
- No error is returned.

---

### Test: Directory with only subdirectories

**Setup:**
- Create a temporary directory within the project root.
- Inside that directory, create one or more subdirectories, but no files at any level.

**Action:**
- Call `ListFiles` with the `PathCfs` for the temporary directory.

**Expected outcome:**
- Returns an empty list.
- No error is returned.

---

## Failure Cases

### Test: Directory does not exist

**Setup:**
- No directory is created. Use a path that is known to not exist on the filesystem.

**Action:**
- Call `ListFiles` with a `PathCfs` pointing to the non-existent directory.

**Expected outcome:**
- Returns error `DirectoryNotFound`.

---

### Test: Propagates validation errors from PathCfsToOs

**Setup:**
- No directory setup required.

**Action:**
- Call `ListFiles` with an invalid `PathCfs` value such as `"../../outside"` that
  would escape the project root.

**Expected outcome:**
- Returns error `DirectoryTraversal` (propagated from PathUtils).
- The function does not attempt any filesystem traversal.

---

### Test: Propagates conversion errors from PathOsToCfs

**Skip condition:** Skip this test on platforms where symlinks are not supported.

**Setup:**
- Create a temporary directory within the project root.
- Inside that directory, create a regular file.
- Inside that directory, create a symlink that points to a file located outside
  the project root.

**Action:**
- Call `ListFiles` with the `PathCfs` for the temporary directory.

**Expected outcome:**
- Returns error `ResolvesOutsideRoot` (propagated from PathUtils).

---

### Test: Walk error

**Skip condition:** Skip this test on platforms where directory permissions cannot
prevent traversal (e.g., when running as a superuser).

**Setup:**
- Create a temporary directory within the project root.
- Inside that directory, create a subdirectory.
- Inside the subdirectory, create at least one file.
- Set the permissions on the subdirectory to deny read access.

**Action:**
- Call `ListFiles` with the `PathCfs` for the parent temporary directory.

**Expected outcome:**
- Returns error `WalkError`.

**Teardown note:**
- Restore permissions on the subdirectory so it can be cleaned up after the test.
