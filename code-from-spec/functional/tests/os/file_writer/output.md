<!-- code-from-spec: SPEC/functional/tests/os/file_writer@TO61LggAIIUBGzYP7sZVi_skxm0 -->

## Test suite: FileWrite

---

### TC-01: Writes content to a new file

Setup:
- Ensure the target file does not exist.

Actions:
1. Call `FileWrite` with a valid path to a non-existent file and content `"hello world"`.

Expected outcome:
- No error is returned.
- The file exists at the resolved path.
- The file content is exactly `"hello world"`.

---

### TC-02: Overwrites an existing file

Setup:
- Create a file at the target path with content `"old"`.

Actions:
1. Call `FileWrite` with the same path and content `"new"`.

Expected outcome:
- No error is returned.
- The file content is exactly `"new"`.
- No trace of `"old"` remains.

---

### TC-03: Creates intermediate directories

Setup:
- Ensure the parent directories (e.g., `"a/b/c"`) do not exist.

Actions:
1. Call `FileWrite` with a path whose parent directories do not exist (e.g., `"a/b/c/file.txt"`) and any non-empty content.

Expected outcome:
- No error is returned.
- All intermediate directories are created.
- The file exists at the resolved path with the given content.

---

### TC-04: Preserves UTF-8 content

Setup:
- Ensure the target file does not exist.

Actions:
1. Call `FileWrite` with a valid path and content `"café 日本語 🎉"`.
2. Read the file back as raw bytes.

Expected outcome:
- No error is returned.
- The raw bytes of the file match the UTF-8 encoding of `"café 日本語 🎉"` byte-for-byte.

---

### TC-05: Preserves line endings as received

Setup:
- Ensure the target file does not exist.

Actions:
1. Call `FileWrite` with a valid path and content `"alpha\r\nbeta\r\n"`.
2. Read the file back as raw bytes.

Expected outcome:
- No error is returned.
- The raw bytes contain CRLF sequences (`\r\n`) exactly as provided — no normalization has occurred.

---

### TC-06: Writes empty content

Setup:
- Ensure the target file does not exist.

Actions:
1. Call `FileWrite` with a valid path and an empty string as content.

Expected outcome:
- No error is returned.
- The file exists at the resolved path.
- The file contains zero bytes.

---

### TC-07: Propagates validation errors from PathCfsToOs

Setup:
- No files or directories are created.

Actions:
1. Call `FileWrite` with an invalid `PathCfs` value such as `"../../outside"`.

Expected outcome:
- A `DirectoryTraversal` error (propagated from `PathUtils`) is returned.
- No file is created.
- No directory is created.

---

### TC-08: Cannot create directory

Setup:
- Create a regular file at a path that would need to be an intermediate directory (e.g., a file named `"a"` where the target path is `"a/b/file.txt"`).

Actions:
1. Call `FileWrite` with a path whose intermediate directory component conflicts with the existing file.

Expected outcome:
- A `CannotCreateDirectory` error is returned.
- No new file is created.

---

### TC-09: Cannot write file

Setup:
- Create a directory at the exact target path (e.g., a directory named `"file.txt"`).

Actions:
1. Call `FileWrite` with a path that resolves to an existing directory.

Expected outcome:
- A `CannotWriteFile` error is returned.
- The directory remains unchanged.
