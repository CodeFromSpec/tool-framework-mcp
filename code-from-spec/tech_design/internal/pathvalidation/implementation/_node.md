---
depends_on:
  - ROOT/external/owasp-path-traversal
outputs:
  - id: pathvalidation
    path: internal/pathvalidation/pathvalidation.go
---

# ROOT/tech_design/internal/pathvalidation/implementation

Generates the pathvalidation package implementation.

# Agent

## Implementation

1. Reject empty paths.
2. Reject absolute paths. Use `strings.HasPrefix(path, "/")` to
   catch Unix-style absolute paths (including on Windows, where
   `filepath.IsAbs` returns false for paths starting with `/`
   without a drive letter). Also reject if the path contains `:`
   (Windows drive letter, e.g. `C:\...`).
3. Call `filepath.Clean` on the path to normalize separators
   and resolve `.` segments.
4. Reject if any component is `..` after cleaning.
5. Resolve the full absolute path:
   `abs := filepath.Join(projectRoot, cleaned)`.
6. Call `filepath.EvalSymlinks` on `abs` to resolve any
   symlinks in the path. If the target does not exist yet,
   evaluate the longest existing prefix.
7. Verify that the resolved path starts with `projectRoot`.
   If not, the path escapes the project — reject it.

## Constraints

- This function must not write or create anything on disk.
  It is read-only validation.
- Never attempt to sanitize or fix an invalid path. Reject
  and report — the caller decides what to do.
- Every error must identify the offending path.
