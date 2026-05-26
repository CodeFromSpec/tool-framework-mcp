<!-- code-from-spec: ROOT/functional/utils/path_validation@BcOZCM6XsELXQa3-mXNc6aT71r4 -->

# Path Validation

## Overview

Given a relative file path and a project root directory, determines whether
the path is safe to use. Returns success or an error describing the violation.

This function is read-only: it never creates, modifies, or deletes any file.
It never sanitizes input — it rejects invalid paths outright.
Every error message identifies the offending path.

---

## Functions

---

### function ValidatePath(relative_path, project_root) -> void

**Parameters**

- `relative_path` — string — the file path to validate, as provided by the caller
- `project_root` — string — the absolute path of the project root directory

**Returns**

- nothing on success

**Errors**

- `"path is empty: <relative_path>"` — the path string is empty
- `"path is absolute: <relative_path>"` — the path starts with `/` or a drive letter (e.g. `C:`)
- `"directory traversal: <relative_path>"` — the path contains `..` components after normalization
- `"resolves outside root: <relative_path>"` — after resolving symlinks, the path falls outside the project root

---

#### Logic

1. **Reject empty paths.**
   If `relative_path` is an empty string, raise error `"path is empty: <relative_path>"`.

2. **Reject absolute paths.**
   If `relative_path` starts with `/`,
   or matches the pattern of a drive letter followed by `:` (e.g. `C:`, `D:`),
   raise error `"path is absolute: <relative_path>"`.

3. **Normalize the path.**
   Apply OS-aware path normalization to `relative_path`:
   - Replace all backslash separators with the canonical separator.
   - Collapse any `.` components (current directory references).
   - Collapse any `..` components against preceding components where possible.
   - Remove duplicate or trailing separators.
   Store the result as `normalized_path`.

4. **Reject remaining `..` components.**
   Split `normalized_path` into individual path components.
   For each component:
     If the component equals `..`,
     raise error `"directory traversal: <relative_path>"`.

   Note: after step 3, a leading `..` that could not be collapsed (because
   there is no preceding component to consume it) will survive as a `..`
   component and must be caught here.

5. **Resolve the absolute path.**
   Join `project_root` and `normalized_path` to produce `absolute_path`.
   This is a string join using the OS path separator — no filesystem access yet.

6. **Resolve symlinks.**
   Resolve all symbolic links in `absolute_path` by following the filesystem.
   Store the fully resolved, real path as `real_path`.
   If the filesystem cannot resolve the path (e.g. the file does not exist),
   use the canonical form of `absolute_path` without symlink resolution.

7. **Verify the resolved path is within the project root.**
   Resolve symlinks in `project_root` as well, producing `real_root`.
   Check that `real_path` starts with `real_root` followed by a path separator,
   or equals `real_root` exactly.
   If neither condition holds, raise error `"resolves outside root: <relative_path>"`.

8. **Return success.**
   The path is safe. Return without error.

---

## Error conditions summary

| Condition | Trigger | Error message |
|---|---|---|
| Empty path | `relative_path` is `""` | `"path is empty: <relative_path>"` |
| Absolute path | Starts with `/` or drive letter | `"path is absolute: <relative_path>"` |
| Directory traversal | `..` component survives normalization | `"directory traversal: <relative_path>"` |
| Escapes project root | Resolved real path is outside real root | `"resolves outside root: <relative_path>"` |

---

## Threat model notes

- **Relative traversal** (`../../etc/passwd`): caught by step 4 — `..` components remain after normalization.
- **Embedded traversal** (`internal/../../outside/file.go`): step 3 collapses `..` against `internal`, leaving `outside/file.go`; step 7 then catches the escape.
- **OS-specific separators** (backslash on Windows): step 3 normalizes separators before any component analysis.
- **Encoding tricks** (URL-encoded or Unicode sequences): callers are responsible for decoding before passing to this function; this function operates on the decoded string value as-is.
- **Symlinks**: step 6 resolves symlinks before the boundary check in step 7, preventing a symlink inside the project from pointing outside it.
