<!-- code-from-spec: ROOT/functional/utils/path_validation@UsR24-uUTsFPeW3dwOyFoQa4PUA -->

# PathValidation

Determines whether a relative file path is safe to use within a project root.
Returns success or an error describing the violation. Never modifies files or
sanitizes input — invalid paths are rejected outright.

---

## function ValidatePath(relative_path, project_root) -> void

**Parameters**
- `relative_path` — string: the path to validate, as provided by the caller
- `project_root` — string: absolute path to the root of the project

**Returns**
- nothing on success

**Errors**
- `"path is empty: <relative_path>"` — the path string is empty
- `"path is absolute: <relative_path>"` — the path starts with `/` or a drive letter like `C:`
- `"directory traversal: <relative_path>"` — a `..` component remains after normalization
- `"resolves outside root: <relative_path>"` — after resolving symlinks, the final path is outside the project root

---

### Steps

1. **Reject empty paths.**
   If `relative_path` is an empty string, raise error `"path is empty: <relative_path>"`.

2. **Reject absolute paths.**
   Check whether `relative_path` begins with `/`.
   Also check whether `relative_path` matches the pattern of a drive letter followed by `:`
   (for example `C:`, `D:`, etc.), which identifies an absolute Windows path.
   If either condition is true, raise error `"path is absolute: <relative_path>"`.

3. **Normalize separators.**
   Replace all backslash (`\`) characters in `relative_path` with forward slash (`/`).
   This neutralizes OS-specific separator variations.

4. **Decode any percent-encoded or Unicode escape sequences.**
   If the path contains URL-encoded sequences (e.g., `%2F`, `%2E`) or Unicode escape
   sequences, decode them to their canonical character representation.
   This prevents encoding tricks that disguise traversal characters.

5. **Normalize `.` and `..` components.**
   Apply a standard path normalization that resolves `.` (current directory) and
   `..` (parent directory) components lexically, without touching the filesystem.

6. **Reject remaining `..` components.**
   After normalization, inspect each path component by splitting on `/`.
   If any component equals `..`, raise error `"directory traversal: <relative_path>"`.
   Note: use the original `relative_path` value in the error message, not the
   normalized form, so the caller can identify what was submitted.

7. **Build the candidate absolute path.**
   Join `project_root` and the normalized path to form a single absolute path.
   Ensure exactly one separator exists at the join point.

8. **Resolve symlinks.**
   Follow all symbolic links in the candidate absolute path to obtain the real,
   physical path on the filesystem.
   If any segment of the path does not exist, the symlink resolution cannot be
   completed — raise error `"resolves outside root: <relative_path>"`.

9. **Verify the resolved path is within the project root.**
   Resolve symlinks in `project_root` as well, to obtain its canonical form.
   Check that the resolved candidate path begins with the canonical project root,
   followed by a separator or end-of-string (to avoid a prefix match like
   `/project-extra` being accepted for root `/project`).
   If the check fails, raise error `"resolves outside root: <relative_path>"`.

10. **Return success.**
    The path is safe. Return without error.

---

### Contracts

- **Read-only.** This function never creates, writes, or modifies any file.
- **No sanitization.** Every invalid input is rejected with an error; no path
  is silently rewritten or truncated.
- **Every error identifies the offending path.** The original `relative_path`
  value must appear in every error message to aid debugging.
