---
depends_on:
  - ROOT/functional/dependencies/owasp-path-traversal
input: ARTIFACT/functional/logic/utils/path_validation(path_validation)
outputs:
  - id: pathvalidation
    path: internal/pathvalidation/pathvalidation.go
---

# ROOT/golang/implementation/internal/pathvalidation/code

Generates the pathvalidation package implementation.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `filepath.Clean`, `filepath.Join`, `filepath.EvalSymlinks`,
  and `filepath.ToSlash` from the standard library.
- Use `strings.HasPrefix(path, "/")` to catch Unix-style absolute
  paths (including on Windows, where `filepath.IsAbs` returns
  false for paths starting with `/` without a drive letter).
  Also reject if the path contains `:` (Windows drive letter).
- The package name should be `pathvalidation`.
- This function must not write or create anything on disk.
  It is read-only validation.
- Never attempt to sanitize or fix an invalid path. Reject
  and report -- the caller decides what to do.
- Every error must identify the offending path.
