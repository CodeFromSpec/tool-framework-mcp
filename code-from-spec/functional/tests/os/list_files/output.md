<!-- code-from-spec: ROOT/functional/tests/os/list_files@Eqkl9ohPw0ZEvY_XYzaG_oJTcH4 -->

# Test Specification: ListFiles

```
function ListFiles(cfs_path: PathCfs) -> list of PathCfs
  errors:
    - DirectoryNotFound: the directory does not exist.
    - WalkError: a filesystem error occurred while traversing.
    - (PathUtils.*): propagated from PathCfsToOs.
    - (PathUtils.*): propagated from PathOsToCfs.
```

---

## Happy Path

### TC-01: Lists files in a flat directory

**Setup:**
Create a temporary directory containing three files: `a.txt`, `b.txt`, `c.txt`.

**Action:**
Call `ListFiles` with the path to that directory.

**Expected outcome:**
Returns a list of three `PathCfs` values in alphabetical order:
- `<dir>/a.txt`
- `<dir>/b.txt`
- `<dir>/c.txt`

(each relative to the project root)

No error is returned.

---

### TC-02: Lists files recursively

**Setup:**
Create a temporary directory `dir` with the following structure:
```
dir/
  alpha.txt
  sub/
    beta.txt
    deep/
      gamma.txt
```

**Action:**
Call `ListFiles` with the path to `dir`.

**Expected outcome:**
Returns a list of three `PathCfs` values in alphabetical order:
- `dir/alpha.txt`
- `dir/sub/beta.txt`
- `dir/sub/deep/gamma.txt`

(each relative to the project root)

No error is returned.

---

### TC-03: Results are sorted alphabetically

**Setup:**
Create a temporary directory containing three files: `z.txt`, `a.txt`, `m.txt`.

**Action:**
Call `ListFiles` with the path to that directory.

**Expected outcome:**
Returns a list of three `PathCfs` values in alphabetical order:
- `<dir>/a.txt`
- `<dir>/m.txt`
- `<dir>/z.txt`

No error is returned.

---

## Edge Cases

### TC-04: Empty directory

**Setup:**
Create a temporary directory containing no files and no subdirectories.

**Action:**
Call `ListFiles` with the path to that directory.

**Expected outcome:**
Returns an empty list. No error is returned.

---

### TC-05: Directory with only subdirectories

**Setup:**
Create a temporary directory containing only subdirectories (no files at any level).

**Action:**
Call `ListFiles` with the path to that directory.

**Expected outcome:**
Returns an empty list. No error is returned.

---

## Failure Cases

### TC-06: Directory does not exist

**Setup:**
No setup required.

**Action:**
Call `ListFiles` with a `PathCfs` that refers to a directory that does not exist on the filesystem.

**Expected outcome:**
Returns error `DirectoryNotFound`.

---

### TC-07: Propagates validation errors from PathCfsToOs

**Setup:**
No setup required.

**Action:**
Call `ListFiles` with an invalid `PathCfs` value such as `"../../outside"` that attempts to traverse outside the project root.

**Expected outcome:**
Returns error `DirectoryTraversal` (propagated from PathUtils). No files are listed.

---

### TC-08: Propagates conversion errors from PathOsToCfs

**Skip condition:** Skip on platforms where symlinks are not supported.

**Setup:**
Create a temporary directory containing:
- A regular file.
- A symlink that resolves to a file located outside the project root.

**Action:**
Call `ListFiles` with the path to that directory.

**Expected outcome:**
Returns error `ResolvesOutsideRoot` (propagated from PathUtils).

---

### TC-09: Walk error

**Skip condition:** Skip on platforms where directory permissions cannot prevent traversal (e.g., when running as root).

**Setup:**
Create a temporary directory containing a subdirectory. Set the permissions on the subdirectory so that it cannot be read or entered.

**Action:**
Call `ListFiles` with the path to the parent directory.

**Expected outcome:**
Returns error `WalkError`.
