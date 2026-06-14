<!-- code-from-spec: ROOT/functional/tests/os/file_writer@4gZURUXGii3sDorMietZne9KMCc -->

## Test: Writes content to a new file

Setup:
- Prepare a valid `PathCfs` pointing to a file that does not exist.

Actions:
1. Call `FileWrite` with that path and content `"hello world"`.

Expected outcome:
- No error is returned.
- The file exists at the resolved OS path.
- The file content is exactly `"hello world"`.

---

## Test: Overwrites an existing file

Setup:
- Prepare a valid `PathCfs` pointing to a file location.
- Create a file at that location with content `"old"`.

Actions:
1. Call `FileWrite` with the same path and content `"new"`.

Expected outcome:
- No error is returned.
- The file content is exactly `"new"`.
- No trace of `"old"` remains.

---

## Test: Creates intermediate directories

Setup:
- Prepare a valid `PathCfs` whose parent directories do not exist
  (e.g., resolving to `"a/b/c/file.txt"` where `a`, `b`, and `c`
  are absent).

Actions:
1. Call `FileWrite` with that path and content `"hello"`.

Expected outcome:
- No error is returned.
- All intermediate directories are created.
- The file exists with content `"hello"`.

---

## Test: Preserves UTF-8 content

Setup:
- Prepare a valid `PathCfs` pointing to a file that does not exist.

Actions:
1. Call `FileWrite` with content `"café 日本語 🎉"`.
2. Read the file back as raw bytes.

Expected outcome:
- No error is returned.
- The raw bytes match the UTF-8 encoding of `"café 日本語 🎉"` exactly.

---

## Test: Preserves line endings as received

Setup:
- Prepare a valid `PathCfs` pointing to a file that does not exist.

Actions:
1. Call `FileWrite` with content `"alpha\r\nbeta\r\n"`.
2. Read the file back as raw bytes.

Expected outcome:
- No error is returned.
- The raw bytes contain CRLF sequences unchanged — no normalization
  to LF or any other transformation.

---

## Test: Writes empty content

Setup:
- Prepare a valid `PathCfs` pointing to a file that does not exist.

Actions:
1. Call `FileWrite` with an empty string as content.

Expected outcome:
- No error is returned.
- The file exists with zero bytes.

---

## Test: Propagates validation errors from PathCfsToOs

Setup:
- Prepare a `PathCfs` value that is invalid due to directory
  traversal (e.g., `"../../outside"`).
- Note the state of the filesystem before the call.

Actions:
1. Call `FileWrite` with the invalid path and content `"x"`.

Expected outcome:
- Error DirectoryTraversal is returned (propagated from PathUtils).
- No file is created.
- No directory is created.
- The filesystem is unchanged from the state noted in setup.

---

## Test: Cannot create directory

Setup:
- Prepare a valid `PathCfs` whose resolved OS path contains an
  intermediate path component that conflicts with an existing file
  (e.g., a file named `"a"` already exists, and the path resolves
  to `"a/b/file.txt"`).

Actions:
1. Call `FileWrite` with that path and content `"x"`.

Expected outcome:
- Error CannotCreateDirectory is returned.

---

## Test: Cannot write file

Setup:
- Prepare a valid `PathCfs` whose resolved OS path points to a
  directory that already exists (not a file).

Actions:
1. Call `FileWrite` with that path and content `"x"`.

Expected outcome:
- Error CannotWriteFile is returned.
