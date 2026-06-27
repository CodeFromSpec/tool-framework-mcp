---
depends_on:
  - SPEC/functional/dependencies/owasp-path-traversal
output: code-from-spec/functional/logic/os/path_utils/output.md
---

# SPEC/functional/logic/os/path_utils

Path types and safe path conversion for the framework.

# Public

## Namespace

    namespace: pathutils

## Interface

```
record PathCfs
  value: string

record PathOs
  value: string

function PathGetProjectRoot() -> PathOs
  errors:
    - CannotDetermineRoot: the working directory cannot
      be read.

function PathValidateCfs(value: string)
  errors:
    - PathEmpty: the path value is empty.
    - PathAbsolute: the path starts with / or a drive
      letter like C:.
    - PathContainsBackslash: the path contains \
      characters.
    - DirectoryTraversal: the path contains ..
      components after normalization.

function PathCfsToOs(cfs_path: PathCfs) -> PathOs
  errors:
    - ResolvesOutsideRoot: after resolving symlinks,
      the path is outside the project root.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (PathUtils.*): propagated from PathGetProjectRoot.

function PathOsToCfs(os_path: PathOs) -> PathCfs
  errors:
    - ResolvesOutsideRoot: the path is not within the
      project root.
    - (PathUtils.*): propagated from PathGetProjectRoot.
```

### PathCfs

A path in the Code from Spec standard format:
- Forward slash (`/`) as separator, always.
- Relative to the project root.
- No `..` components, no drive letters, no leading `/`,
  no backslashes.

This is the only path format used in the framework's
public API — in frontmatter fields (`output`, `input`),
in logical names, and in tool parameters.

Examples:
- `internal/file/file.go`
- `code-from-spec/functional/logic/os/file/_node.md`

### PathOs

An absolute path in the operating system's native format:
- OS-specific separator (`/` on Unix, `\` on Windows).
- Always absolute.

This type is never exposed in the framework's public API.
It exists only inside the `os/` layer for interacting
with the filesystem.

Examples:
- `/home/user/myproject/internal/file/file.go` (Unix)
- `C:\Users\user\myproject\internal\file\file.go` (Windows)

### PathGetProjectRoot

Returns the project root as a `PathOs`. Determined from
the working directory of the process.

### PathValidateCfs

Validates that a value conforms to the `PathCfs` format
rules. Raises an error describing the violation if not.
Follows OWASP guidance for path traversal prevention.

Use this for sanity checks on parameters received from
callers. Does not verify that the file exists or resolve
symlinks — use `PathCfsToOs` for that.

### PathCfsToOs

Validates a `PathCfs` and converts it to an absolute
`PathOs`. This is the single entry point for going from
framework paths to OS paths. If validation fails, no
conversion happens — an error is returned.

The target file or directory does not need to exist.
The conversion is purely path-based — it validates the
format, converts separators, and checks containment,
but does not require the path to resolve to an actual
filesystem entry.

### PathOsToCfs

Converts an absolute `PathOs` to a `PathCfs` relative to
the project root. Used internally by components that
receive paths from the OS (e.g. directory listing).

The target file or directory does not need to exist.
The conversion is purely path-based.

# Agent

Generate pseudocode for all functions.

## Implementation guidance

### PathValidateCfs steps

1. Raise "path is empty" if the value is empty.
2. Raise "path is absolute" if the value starts with `/`
   or a drive letter (e.g. `C:`).
3. Raise "path contains backslash" if the value contains
   `\`.
4. Normalize the path (resolve `.` and `..`).
5. Raise "directory traversal" if any component is `..`
   after normalization.

### PathCfsToOs steps

1. Call `PathValidateCfs`. If it raises an error,
   propagate it.
2. Replace forward slashes with the OS separator.
3. Join with the project root to form an absolute path.
4. If the path exists on disk, resolve symlinks and
   verify the resolved path is within the project root.
   Raise "resolves outside root" if not.
5. Return the absolute PathOs.

### PathOsToCfs steps

1. If the path exists on disk, resolve symlinks.
2. Verify the path is within the project root. Raise
   "resolves outside root" if not.
3. Compute the relative path from project root.
4. Replace OS separators with forward slashes.
5. Return the PathCfs.

### PathGetProjectRoot

Returns the working directory of the process as a
`PathOs`.

### Security

See `SPEC/functional/dependencies/owasp-path-traversal`
for the threat model. PathCfsToOs is the primary defense
against path traversal attacks.

## Contracts

- PathCfsToOs never creates or modifies files.
- PathCfsToOs never sanitizes — rejects invalid paths.
- PathOsToCfs never creates or modifies files.
- Both conversion functions resolve symlinks before
  checking containment.
