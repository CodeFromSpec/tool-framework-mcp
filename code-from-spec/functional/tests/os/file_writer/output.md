<!-- code-from-spec: ROOT/functional/tests/os/file_writer@rQPjqMZy8Q8qTGQEBjPHBv4WvrU -->

# Test Specification: FileWrite

## Interface

```
function FileWrite(cfs_path: pathutils.PathCfs, content: string)
  errors:
    - CannotCreateDirectory
    - CannotWriteFile
    - (PathUtils.*): propagated from PathCfsToOs
```

---

## Happy Path

### TC-1: Writes content to a new file

**Setup:**
- Use a `cfs_path` that points to a file that does not exist.

**Actions:**
1. Call `FileWrite` with the chosen `cfs_path` and content `"hello world"`.

**Expected outcome:**
- No error is returned.
- The file is created at the resolved path.
- The file content is exactly `"hello world"`.

---

### TC-2: Overwrites an existing file

**Setup:**
- Create a file at a known path with content `"old"`.

**Actions:**
1. Call `FileWrite` with the same `cfs_path` and content `"new"`.

**Expected outcome:**
- No error is returned.
- The file content is exactly `"new"`.
- The old content `"old"` is completely replaced.

---

### TC-3: Creates intermediate directories

**Setup:**
- Use a `cfs_path` whose parent directories do not exist
  (e.g., a path equivalent to `"a/b/c/file.txt"` where `a`, `b`, and `c` are absent).

**Actions:**
1. Call `FileWrite` with the chosen `cfs_path` and any non-empty content.

**Expected outcome:**
- No error is returned.
- All intermediate directories (`a`, `a/b`, `a/b/c`) are created.
- The file is created at the resolved path with the given content.

---

### TC-4: Preserves UTF-8 content

**Setup:**
- Use a `cfs_path` that points to a file that does not exist.

**Actions:**
1. Call `FileWrite` with the chosen `cfs_path` and content `"café 日本語 🎉"`.
2. Read the file back from disk as raw bytes.

**Expected outcome:**
- No error is returned.
- The raw bytes of the file match the UTF-8 encoding of `"café 日本語 🎉"` byte-for-byte.

---

### TC-5: Preserves line endings as received

**Setup:**
- Use a `cfs_path` that points to a file that does not exist.

**Actions:**
1. Call `FileWrite` with the chosen `cfs_path` and content `"alpha\r\nbeta\r\n"`
   (CRLF line endings).
2. Read the file back from disk as raw bytes.

**Expected outcome:**
- No error is returned.
- The raw bytes of the file contain CRLF sequences (`\r\n`) — no normalization has occurred.

---

### TC-6: Writes empty content

**Setup:**
- Use a `cfs_path` that points to a file that does not exist.

**Actions:**
1. Call `FileWrite` with the chosen `cfs_path` and content `""` (empty string).

**Expected outcome:**
- No error is returned.
- The file is created at the resolved path.
- The file contains zero bytes.

---

## Failure Cases

### TC-7: Propagates validation errors from PathCfsToOs

**Setup:**
- Use an invalid `cfs_path` that would trigger a directory traversal violation
  (e.g., a path containing `"../../outside"`).

**Actions:**
1. Call `FileWrite` with the invalid `cfs_path` and any content.

**Expected outcome:**
- Error `DirectoryTraversal` (from PathUtils) is returned.
- No file is created.
- No intermediate directory is created.

---

### TC-8: Cannot create directory

**Setup:**
- Create a regular file at a path that would conflict with an intermediate directory
  (e.g., place a file named `"a"` where `FileWrite` would need to create a directory
  named `"a"` to satisfy a path like `"a/b/file.txt"`).

**Actions:**
1. Call `FileWrite` with a `cfs_path` that requires creating the conflicting directory.

**Expected outcome:**
- Error `CannotCreateDirectory` is returned.
- No target file is created.

---

### TC-9: Cannot write file

**Setup:**
- Create a directory at the exact path where `FileWrite` would write a file
  (e.g., create directory `"mydir"` and use a `cfs_path` that resolves to `"mydir"`).

**Actions:**
1. Call `FileWrite` with a `cfs_path` that resolves to the existing directory.

**Expected outcome:**
- Error `CannotWriteFile` is returned.
- The existing directory is not modified.
