<!-- code-from-spec: ROOT/functional/tests/os/list_files@8a__BUD_FR1ccWj6DOUOPwrZVLM -->

# Test Specification: ListFiles

## Interface

```
function ListFiles(cfs_path: PathCfs) -> list of PathCfs
```

---

## Test Cases

### Happy Path

---

#### TC-01: Lists files in a flat directory

**Setup**
- Create a temporary directory at a valid project-relative path.
- Inside that directory, create three files: `a.txt`, `b.txt`, `c.txt`.

**Action**
- Call `ListFiles` with the path to that directory.

**Expected Outcome**
- Returns a list of three `PathCfs` values.
- The values are, in order: `<dir>/a.txt`, `<dir>/b.txt`, `<dir>/c.txt`
  (paths relative to the project root, sorted alphabetically).
- No error is raised.

---

#### TC-02: Lists files recursively

**Setup**
- Create the following directory structure at a valid project-relative path:
  ```
  dir/
    alpha.txt
    sub/
      beta.txt
      deep/
        gamma.txt
  ```

**Action**
- Call `ListFiles` with the path to `dir`.

**Expected Outcome**
- Returns a list of three `PathCfs` values.
- The values are, in order:
  `dir/alpha.txt`, `dir/sub/beta.txt`, `dir/sub/deep/gamma.txt`
  (sorted alphabetically).
- No error is raised.

---

#### TC-03: Results are sorted alphabetically

**Setup**
- Create a temporary directory at a valid project-relative path.
- Inside that directory, create three files in this order: `z.txt`, `a.txt`, `m.txt`.

**Action**
- Call `ListFiles` with the path to that directory.

**Expected Outcome**
- Returns a list of three `PathCfs` values.
- The values are, in order: `<dir>/a.txt`, `<dir>/m.txt`, `<dir>/z.txt`.
- No error is raised.

---

### Edge Cases

---

#### TC-04: Empty directory

**Setup**
- Create a temporary directory at a valid project-relative path.
- Do not create any files or subdirectories inside it.

**Action**
- Call `ListFiles` with the path to that directory.

**Expected Outcome**
- Returns an empty list.
- No error is raised.

---

#### TC-05: Directory with only subdirectories

**Setup**
- Create a temporary directory at a valid project-relative path.
- Inside that directory, create one or more subdirectories but no files at any level.

**Action**
- Call `ListFiles` with the path to that directory.

**Expected Outcome**
- Returns an empty list.
- No error is raised.

---

### Failure Cases

---

#### TC-06: Directory does not exist

**Setup**
- Choose a path that is valid according to `PathCfs` rules but refers to a
  directory that does not exist on disk.

**Action**
- Call `ListFiles` with that path.

**Expected Outcome**
- Raises error "directory not found".

---

#### TC-07: Propagates validation errors from PathCfsToOs

**Setup**
- No filesystem setup required.

**Action**
- Call `ListFiles` with an invalid `PathCfs` value such as `"../../outside"`.

**Expected Outcome**
- Raises error "directory traversal" propagated from `PathCfsToOs`.

---

#### TC-08: Propagates conversion errors from PathOsToCfs

**Note:** Skip this test case on platforms where symlinks are not supported.

**Setup**
- Create a temporary directory at a valid project-relative path.
- Inside that directory, create a regular file.
- Inside that directory, also create a symlink that resolves to a file whose
  absolute path is outside the project root.

**Action**
- Call `ListFiles` with the path to that directory.

**Expected Outcome**
- Raises error "resolves outside root" propagated from `PathOsToCfs`.

---

#### TC-09: Walk error

**Note:** Skip this test case on platforms where directory permissions cannot
prevent traversal (e.g., some Windows configurations).

**Setup**
- Create a temporary directory at a valid project-relative path.
- Inside that directory, create a subdirectory.
- Set the permissions on that subdirectory so that its contents cannot be read.

**Action**
- Call `ListFiles` with the path to the parent directory.

**Expected Outcome**
- Raises error "walk error".
