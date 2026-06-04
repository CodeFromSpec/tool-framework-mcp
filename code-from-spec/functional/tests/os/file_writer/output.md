<!-- code-from-spec: ROOT/functional/tests/os/file_writer@E4aR5VW5Xh5ZGwD5jEogd5lZN9o -->

## Test cases for FileWrite

### Happy path

#### Writes content to a new file

Setup: a path to a file that does not exist.

Actions:
1. Call `FileWrite` with that path and content `"hello world"`.

Expected outcome: no error is returned. The file exists and its content is exactly `"hello world"`.

---

#### Overwrites an existing file

Setup: a file exists at the target path with content `"old"`.

Actions:
1. Call `FileWrite` with the same path and content `"new"`.

Expected outcome: no error is returned. The file content is exactly `"new"`. No trace of the old content remains.

---

#### Creates intermediate directories

Setup: a path whose parent directories do not exist (e.g., `"a/b/c/file.txt"`).

Actions:
1. Call `FileWrite` with that path and any content.

Expected outcome: no error is returned. All intermediate directories are created. The file exists at the target path with the given content.

---

#### Preserves UTF-8 content

Setup: a path to a file that does not exist.

Actions:
1. Call `FileWrite` with content `"café 日本語 🎉"`.
2. Read the file back.

Expected outcome: no error is returned. The bytes read from the file match the UTF-8 encoding of `"café 日本語 🎉"` exactly.

---

#### Preserves line endings as received

Setup: a path to a file that does not exist.

Actions:
1. Call `FileWrite` with content `"alpha\r\nbeta\r\n"`.
2. Read the file back.

Expected outcome: no error is returned. The bytes read from the file contain CRLF line endings — no normalization has occurred.

---

#### Writes empty content

Setup: a path to a file that does not exist.

Actions:
1. Call `FileWrite` with an empty string as content.

Expected outcome: no error is returned. The file exists and contains zero bytes.

---

### Failure cases

#### Propagates validation errors from PathCfsToOs

Setup: an invalid `PathCfs` value, e.g., `"../../outside"`.

Actions:
1. Call `FileWrite` with that path and any content.

Expected outcome: error DirectoryTraversal is returned (propagated from PathUtils). No file or directory is created.

---

#### Cannot create directory

Setup: a path where an intermediate directory cannot be created because a path component conflicts with an existing file (e.g., a file named `"a"` already exists, but the target path is `"a/b/file.txt"`).

Actions:
1. Call `FileWrite` with that path and any content.

Expected outcome: error CannotCreateDirectory is returned. No new directory or file is created.

---

#### Cannot write file

Setup: a path that points to a directory that already exists (not a file).

Actions:
1. Call `FileWrite` with that path and any content.

Expected outcome: error CannotWriteFile is returned.
