<!-- code-from-spec: ROOT/functional/tests/os/file_writer@L9y78mBRM-PVFWj38FzTADuEXaw -->

# FileWrite Test Specification

## Happy Path

### Writes content to a new file

Setup:
- A path that points to a file that does not exist.

Actions:
1. Call `FileWrite` with the path and content `"hello world"`.

Expected outcome:
- No error is returned.
- The file exists at the given path.
- The file content is exactly `"hello world"`.

---

### Overwrites an existing file

Setup:
- A file already exists at the target path with content `"old"`.

Actions:
1. Call `FileWrite` with the same path and content `"new"`.

Expected outcome:
- No error is returned.
- The file content is exactly `"new"`.
- The old content `"old"` is gone entirely.

---

### Creates intermediate directories

Setup:
- A path whose parent directories do not exist
  (e.g., `"a/b/c/file.txt"` where `a`, `b`, and `c` are absent).

Actions:
1. Call `FileWrite` with that path and any non-empty content.

Expected outcome:
- No error is returned.
- All intermediate directories (`a`, `a/b`, `a/b/c`) are created.
- The file exists at the full path with the given content.

---

### Preserves UTF-8 content

Setup:
- A path to a file that does not exist.

Actions:
1. Call `FileWrite` with content `"café 日本語 🎉"`.
2. Read the file back as raw bytes.

Expected outcome:
- No error is returned.
- The raw bytes of the file match the UTF-8 encoding of `"café 日本語 🎉"` byte-for-byte.

---

### Preserves line endings as received

Setup:
- A path to a file that does not exist.

Actions:
1. Call `FileWrite` with content `"alpha\r\nbeta\r\n"` (CRLF line endings).
2. Read the file back as raw bytes.

Expected outcome:
- No error is returned.
- The raw bytes of the file contain CRLF sequences — no normalization has occurred.

---

### Writes empty content

Setup:
- A path to a file that does not exist.

Actions:
1. Call `FileWrite` with an empty string as content.

Expected outcome:
- No error is returned.
- The file exists at the given path.
- The file size is zero bytes.

---

## Failure Cases

### Propagates validation errors from PathCfsToOs

Setup:
- An invalid `PathCfs` value that would escape the base directory,
  e.g., `"../../outside"`.

Actions:
1. Call `FileWrite` with the invalid path and any content.

Expected outcome:
- Error `DirectoryTraversal` is returned (propagated from PathUtils).
- No file is created.
- No directory is created.

---

### Cannot create directory

Setup:
- A path where an intermediate path component conflicts with
  an existing file (e.g., `"a/b/file.txt"` where `"a/b"` already
  exists as a regular file, not a directory).

Actions:
1. Call `FileWrite` with that path and any content.

Expected outcome:
- Error `CannotCreateDirectory` is returned.
- No new file is created at the target path.

---

### Cannot write file

Setup:
- A path that resolves to an existing directory
  (not a file).

Actions:
1. Call `FileWrite` with that path and any content.

Expected outcome:
- Error `CannotWriteFile` is returned.
- The directory is not modified.
