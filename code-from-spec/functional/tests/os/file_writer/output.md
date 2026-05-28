<!-- code-from-spec: ROOT/functional/tests/os/file_writer@0nJhGWugLvOX7gEXetCTCyxWETM -->

# FileWrite — Test Specification

## Happy Path

---

### Test: Writes content to a new file

Setup:
- Ensure the target file does not exist.

Actions:
1. Call `FileWrite` with a valid `cfs_path` pointing to a non-existent file
   and `content` = `"hello world"`.

Expected outcome:
- No error is returned.
- The file is created at the resolved path.
- The file contains exactly `"hello world"`.

---

### Test: Overwrites an existing file

Setup:
- Create a file at the target path with content `"old"`.

Actions:
1. Call `FileWrite` with the same `cfs_path` and `content` = `"new"`.

Expected outcome:
- No error is returned.
- The file exists at the resolved path.
- The file contains exactly `"new"`.
- No trace of `"old"` remains.

---

### Test: Creates intermediate directories

Setup:
- Ensure the parent directories `a/b/c/` do not exist under the CFS root.

Actions:
1. Call `FileWrite` with a `cfs_path` of `"a/b/c/file.txt"` and any non-empty
   `content`.

Expected outcome:
- No error is returned.
- All intermediate directories `a`, `a/b`, and `a/b/c` are created.
- The file `a/b/c/file.txt` exists and contains the given content.

---

### Test: Preserves UTF-8 content

Setup:
- No prior state required.

Actions:
1. Call `FileWrite` with a valid `cfs_path` and `content` = `"café 日本語 🎉"`.
2. Read the bytes of the written file back from disk.

Expected outcome:
- No error is returned.
- The raw bytes of the file match the UTF-8 encoding of `"café 日本語 🎉"` exactly.

---

### Test: Preserves line endings as received

Setup:
- No prior state required.

Actions:
1. Call `FileWrite` with a valid `cfs_path` and `content` = `"alpha\r\nbeta\r\n"`.
2. Read the raw bytes of the written file back from disk.

Expected outcome:
- No error is returned.
- The raw bytes of the file contain CRLF sequences (`\r\n`) exactly as provided.
- No normalization (e.g., conversion to `\n`) has occurred.

---

### Test: Writes empty content

Setup:
- No prior state required.

Actions:
1. Call `FileWrite` with a valid `cfs_path` and `content` = `""` (empty string).

Expected outcome:
- No error is returned.
- The file is created at the resolved path.
- The file has a size of zero bytes.

---

## Failure Cases

---

### Test: Propagates validation errors from PathCfsToOs

Setup:
- No prior state required.

Actions:
1. Call `FileWrite` with an invalid `cfs_path` such as `"../../outside"` that
   would escape the CFS root.

Expected outcome:
- An error `"directory traversal"` is returned, propagated from `PathCfsToOs`.
- No file is created.
- No directory is created.

---

### Test: Cannot create directory

Setup:
- Create a regular file at a path that will conflict with a required intermediate
  directory (e.g., create a file named `"a"` so that `"a/b/file.txt"` cannot
  be created because `"a"` is not a directory).

Actions:
1. Call `FileWrite` with a `cfs_path` whose intermediate directory component
   conflicts with the existing file (e.g., `"a/b/file.txt"`).

Expected outcome:
- An error `"cannot create directory"` is returned.

---

### Test: Cannot write file

Setup:
- Create a directory at the target path (not a file).

Actions:
1. Call `FileWrite` with a `cfs_path` that resolves to an existing directory.

Expected outcome:
- An error `"cannot write file"` is returned.
