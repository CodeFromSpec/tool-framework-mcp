---
depends_on:
  - SPEC/domain/owasp-path-traversal
output: internal/pathutils/pathutils.go
---

# SPEC/golang/implementation/os/path_utils

Path types and safe path conversion for the framework.

# Public

## Package

`package pathutils`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/pathutils"`

## Interface

```go
type PathCfs struct {
	Value string
}

type PathOs struct {
	Value string
}

func PathGetProjectRoot() (PathOs, error)
func PathValidateCfs(value string) error
func PathCfsToOs(cfsPath PathCfs) (PathOs, error)
func PathOsToCfs(osPath PathOs) (PathCfs, error)
```

### PathCfs

A path in the Code from Spec standard format:
- Forward slash as separator, always.
- Relative to the project root.
- No `..` components, no drive letters, no leading `/`,
  no backslashes.

### PathOs

An absolute path in the OS's native format. Never
exposed in the framework's public API.

### Errors

- `ErrCannotDetermineRoot` (PathGetProjectRoot)
- `ErrPathEmpty`, `ErrPathAbsolute`,
  `ErrPathContainsBackslash`, `ErrDirectoryTraversal`
  (PathValidateCfs)
- `ErrResolvesOutsideRoot` (PathCfsToOs, PathOsToCfs)
- Propagated from PathValidateCfs, PathGetProjectRoot.

# Agent

Implement the `pathutils` package, including its
interface.

## Logic

### PathGetProjectRoot

1. Read the current working directory of the process.
   If it cannot be read, raise error "cannot determine
   project root".
2. Return the working directory as a PathOs.

### PathValidateCfs

1. If value is empty, raise "path is empty".
2. If value starts with "/" or matches a drive letter
   pattern (e.g. "C:"), raise "path is absolute".
3. If value contains "\\", raise "path contains
   backslash".
4. Normalize the path by resolving "." and ".."
   components.
5. For each component in the normalized path, if the
   component is "..", raise "directory traversal
   detected".

### PathCfsToOs

1. Call PathValidateCfs with cfs_path.value. If it
   raises an error, propagate it.
2. Call PathGetProjectRoot. If it raises an error,
   propagate it. Store the result as root.
3. Replace all "/" characters in cfs_path.value with
   the OS-native path separator. Store as os_relative.
4. Join root.value and os_relative to form an absolute
   path. Store as absolute_path.
5. If absolute_path exists on disk, resolve symlinks
   to get resolved_path. If resolved_path does not
   start with root.value, raise "resolves outside root".
   Set absolute_path to resolved_path.
6. Return absolute_path as a PathOs.

### PathOsToCfs

1. Call PathGetProjectRoot. If it raises an error,
   propagate it. Store as root.
2. If os_path.value exists on disk, resolve symlinks.
   Set os_path.value to resolved_path.
3. If os_path.value does not start with root.value,
   raise "resolves outside root".
4. Compute the relative portion by removing root.value
   prefix and any leading path separator. Store as
   relative_path.
5. Replace all OS-native separators in relative_path
   with "/". Store as cfs_value.
6. Return cfs_value as a PathCfs.

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
