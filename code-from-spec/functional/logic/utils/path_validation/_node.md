---
outputs:
  - id: path_validation
    path: code-from-spec/functional/logic/utils/path_validation/output.md
---

# ROOT/functional/logic/utils/path_validation

Validates that a file path is safe to use within a project
directory. Prevents path traversal attacks.

# Public

## Interface

```
function ValidatePath(relative_path, project_root) -> void
  errors:
    - path is empty: the path string is empty.
    - path is absolute: the path starts with / or a drive letter like C:.
    - directory traversal: the path contains .. components after normalization.
    - resolves outside root: after resolving symlinks, the path is outside the project root.
```

`project_root` must be an absolute path. Passing a relative
path (e.g. `"."`) produces undefined behavior — the
containment check in step 7 requires both the candidate and
the root to be in absolute form for a reliable comparison.

# Agent

## Behavior

Given a relative file path and a project root directory,
determines whether the path is safe. Returns success or an
error describing the violation.

### Threat model

- Relative traversal: `../../etc/passwd`
- Embedded traversal: `internal/../../outside/file.go`
- OS-specific separators: backslash on Windows
- Encoding tricks: URL-encoded or Unicode sequences
- Symlinks: a path within the project that resolves outside
  it via a symbolic link

### Validation steps

1. Reject empty paths.
2. Reject absolute paths (leading `/` or drive letter like `C:`).
3. Normalize the path (resolve `.` and `..`, normalize separators).
4. Reject if any component is `..` after normalization.
5. Resolve the full absolute path by joining with project root.
6. Resolve symlinks in the path.
7. Verify the resolved path is within the project root.

## Contracts

- Read-only — never creates or modifies files.
- Never sanitizes — rejects invalid paths outright.
- Every error identifies the offending path.
