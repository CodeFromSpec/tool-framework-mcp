<!-- code-from-spec: ROOT/functional/tests/os/file_writer@-p8-ZDC9mFSXL8bz3Fo5cXWmOYY -->

# Test Specification: FileWriter

## Happy path

### Writes content to a new file

Setup: No file exists at the target path.

Actions:
1. Call `FileWrite` with a path to a non-existent file and content `"hello world"`.
2. Read the file back.
3. Expect the file to exist and contain exactly `"hello world"`.

---

### Overwrites an existing file

Setup: Create a file at the target path with content `"old"`.

Actions:
1. Call `FileWrite` with the same path and content `"new"`.
2. Read the file back.
3. Expect the file to contain `"new"` — the old content is fully replaced.

---

### Creates intermediate directories

Setup: Ensure the parent directories of the target path do not exist (e.g., `"a/b/c/file.txt"` where `a/b/c/` is absent).

Actions:
1. Call `FileWrite` with the nested path and any content.
2. Expect the file and all intermediate directories to be created.
3. Read the file back and confirm the content matches what was written.

---

### Preserves UTF-8 content

Setup: No file exists at the target path.

Actions:
1. Call `FileWrite` with content `"café 日本語 🎉"`.
2. Read the file back as raw bytes.
3. Expect the content to match the original string byte-for-byte in UTF-8 encoding.

---

### Preserves line endings as received

Setup: No file exists at the target path.

Actions:
1. Call `FileWrite` with content `"alpha\r\nbeta\r\n"` (CRLF line endings).
2. Read the file back as raw bytes.
3. Expect the content to contain CRLF — no normalization has occurred.

---

### Writes empty content

Setup: No file exists at the target path.

Actions:
1. Call `FileWrite` with an empty string as content.
2. Expect the file to be created with zero bytes.

---

## Failure cases

### Propagates validation errors from PathCfsToOs

Setup: No file or directory is created before the call.

Actions:
1. Call `FileWrite` with an invalid `PathCfs` such as `"../../outside"`.
2. Expect error DirectoryTraversal (propagated from PathUtils).
3. Confirm that no file or directory was created.

---

### Cannot create directory

Setup: Create a regular file at a path that would otherwise be used as an intermediate directory (e.g., create `"a/b"` as a file, then attempt to write to `"a/b/c/file.txt"`).

Actions:
1. Call `FileWrite` with the conflicting path.
2. Expect error CannotCreateDirectory.

---

### Cannot write file

Setup: Create a directory at the exact path that `FileWrite` would use as the file target (e.g., create directory `"a/b/target"` and then call `FileWrite` with path `"a/b/target"`).

Actions:
1. Call `FileWrite` with the path pointing to an existing directory.
2. Expect error CannotWriteFile.
