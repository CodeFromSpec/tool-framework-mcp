---
depends_on:
  - SPEC/domain/owasp-path-traversal
output: internal/oslayer/path.go
---

# SPEC/golang/implementation/oslayer/path

Implements path types and safe path conversion for
the framework.

# Agent

Implement the path types and functions listed in the
Ownership section as a Go file in package `oslayer`.

## Ownership

This file declares and implements:
- Types: `CfsPath`, `OsPath`
- Functions: `GetProjectRoot`, `ValidateStringIsCfsPath`,
  `CfsPathToOs`, `OsPathToCfs`

The following exist in other files of this package and
can be used but must not be redeclared:
- Error sentinels (`ErrCannotDetermineRoot`,
  `ErrPathEmpty`, `ErrPathAbsolute`,
  `ErrPathContainsBackslash`, `ErrDirectoryTraversal`,
  `ErrResolvesOutsideRoot`) — declared in `errors.go`.

To avoid name collisions with other files in this
package, all identifiers you declare beyond the ones
listed in the Ownership section (functions, variables,
types) must use the suffix `Path`.

## Logic

### GetProjectRoot

1. Read the current working directory of the process.
   If it cannot be read, raise ErrCannotDetermineRoot.
2. Return the working directory as an OsPath.

### ValidateStringIsCfsPath

1. If value is empty, raise ErrPathEmpty.
2. If value starts with "/" or matches a drive letter
   pattern (e.g. "C:"), raise ErrPathAbsolute.
3. If value contains "\\", raise ErrPathContainsBackslash.
4. Normalize the path by resolving "." and ".."
   components.
5. For each component in the normalized path, if the
   component is "..", raise ErrDirectoryTraversal.

### CfsPathToOs

1. Call ValidateStringIsCfsPath with string(cfsPath). If it
   raises an error, propagate it.
2. Call GetProjectRoot. If it raises an error,
   propagate it. Store the result as root.
3. Replace all "/" characters in string(cfsPath) with
   the OS-native path separator. Store as os_relative.
4. Join string(root) and os_relative to form an absolute
   path. Store as absolute_path.
5. If absolute_path exists on disk, resolve symlinks
   to get resolved_path. If resolved_path does not
   start with string(root) followed by the OS path
   separator, and resolved_path is not equal to
   string(root), raise ErrResolvesOutsideRoot.
   Set absolute_path to resolved_path.
6. Return OsPath(absolute_path).

### OsPathToCfs

1. Call GetProjectRoot. If it raises an error,
   propagate it. Store as root.
2. If the path exists on disk, resolve symlinks.
   Let resolved be the result.
3. If resolved does not start with string(root)
   followed by the OS path separator, and resolved is
   not equal to string(root), raise
   ErrResolvesOutsideRoot.
4. Compute the relative portion by removing the root
   prefix and any leading path separator. Store as
   relative_path.
5. Replace all OS-native separators in relative_path
   with "/". Store as cfs_value.
6. Return CfsPath(cfs_value).

## Go-specific guidance

- Use `filepath.Clean`, `filepath.Join`,
  `filepath.EvalSymlinks`, and `filepath.ToSlash` from
  the standard library.
- Use `strings.HasPrefix(path, "/")` to catch Unix-style
  absolute paths (including on Windows, where
  `filepath.IsAbs` returns false for paths starting with
  `/` without a drive letter). Also reject if the path
  contains `:` (Windows drive letter).
- Use `os.Getwd` for GetProjectRoot.
- Read-only — never create or modify files on disk.
- Never sanitize or fix an invalid path — reject and
  report.
