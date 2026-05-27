---
depends_on:
  - ROOT/functional/dependencies/owasp-path-traversal
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/logic/os/path_utils(path_utils)
outputs:
  - id: pathutils
    path: internal/pathutils/pathutils.go
---

# ROOT/golang/implementation/os/path_utils

# Agent

## Go-specific guidance

- Use `filepath.Clean`, `filepath.Join`,
  `filepath.EvalSymlinks`, and `filepath.ToSlash` from
  the standard library.
- Use `strings.HasPrefix(path, "/")` to catch Unix-style
  absolute paths (including on Windows, where
  `filepath.IsAbs` returns false for paths starting with
  `/` without a drive letter). Also reject if the path
  contains `:` (Windows drive letter).
- Use `os.Getwd` for `PathGetProjectRoot`.
- Read-only — never create or modify files on disk.
- Never sanitize or fix an invalid path — reject and
  report.
