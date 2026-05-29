<!-- code-from-spec: ROOT/functional/tests/os/file_writer@0nJhGWugLvOX7gEXetCTCyxWETM -->

# Test Specification: FileWrite

---

## Happy Path

---

### Test: Writes content to a new file

**Setup**
- Ensure the target file does not exist.

**Actions**
1. Call `FileWrite` with a valid `cfs_path` pointing to a non-existent file and content `"hello world"`.

**Expected outcome**
- No error is returned.
- The file exists at the resolved path.
- The file content is exactly `"hello world"`.

---

### Test: Overwrites an existing file

**Setup**
- Create a file at the target path with content `"old"`.

**Actions**
1. Call `FileWrite` with the same `cfs_path` and content `"new"`.

**Expected outcome**
- No error is returned.
- The file content is exactly `"new"`.
- No trace of `"old"` remains in the file.

---

### Test: Creates intermediate directories

**Setup**
- Ensure none of the intermediate directories (e.g., `"a/b/c/"`) exist.

**Actions**
1. Call `FileWrite` with a `cfs_path` whose parent directories do not exist (e.g., `"a/b/c/file.txt"`) and content `"hello"`.

**Expected outcome**
- No error is returned.
- All intermediate directories are created.
- The file exists at the resolved path with content `"hello"`.

---

### Test: Preserves UTF-8 content

**Setup**
- No prior state required.

**Actions**
1. Call `FileWrite` with a valid `cfs_path` and content `"café 日本語 🎉"`.
2. Read the file back from the resolved path.

**Expected outcome**
- No error is returned.
- The file content matches `"café 日本語 🎉"` byte-for-byte.

---

### Test: Preserves line endings as received

**Setup**
- No prior state required.

**Actions**
1. Call `FileWrite` with a valid `cfs_path` and content `"alpha\r\nbeta\r\n"` (CRLF line endings).
2. Read the file back from the resolved path in binary mode.

**Expected outcome**
- No error is returned.
- The file content contains CRLF (`\r\n`) line endings — not normalized to LF.
- The content matches `"alpha\r\nbeta\r\n"` byte-for-byte.

---

### Test: Writes empty content

**Setup**
- No prior state required.

**Actions**
1. Call `FileWrite` with a valid `cfs_path` and an empty string as content.

**Expected outcome**
- No error is returned.
- The file exists at the resolved path.
- The file contains zero bytes.

---

## Failure Cases

---

### Test: Propagates validation errors from PathCfsToOs

**Setup**
- No prior state required.

**Actions**
1. Call `FileWrite` with an invalid `cfs_path` that attempts directory traversal (e.g., `"../../outside"`).

**Expected outcome**
- An error `"directory traversal"` is returned, propagated from `PathCfsToOs`.
- No file is created.
- No directory is created.

---

### Test: Cannot create directory

**Setup**
- Create a regular file at a path that will conflict with a required intermediate directory (e.g., create a file named `"a"`, then attempt to write to `"a/b/file.txt"`).

**Actions**
1. Call `FileWrite` with a `cfs_path` where an intermediate directory component conflicts with an existing file.

**Expected outcome**
- An error `"cannot create directory"` is returned.

---

### Test: Cannot write file

**Setup**
- Create a directory at the target path (e.g., a directory named `"target"` where `"target"` is the intended file path).

**Actions**
1. Call `FileWrite` with a `cfs_path` that resolves to an existing directory (not a file).

**Expected outcome**
- An error `"cannot write file"` is returned.
