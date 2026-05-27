---
depends_on:
  - ROOT/functional/dependencies/owasp-path-traversal
outputs:
  - id: path_utils
    path: code-from-spec/functional/logic/os/path_utils/output.md
---

# ROOT/functional/logic/os/path_utils

Path types and conversion functions for the framework.
Defines the standard path format (CfsPath), the OS-native
format (OsPath), and safe conversions between them.

Review status: pending

# Public

## Interface

```
record CfsPath
  value: string

record OsPath
  value: string

function GetProjectRoot() -> OsPath
  errors:
    - cannot determine root: the working directory cannot
      be read.

function ValidateCfsPath(value) -> boolean

function ResolvePath(cfs_path) -> OsPath
  errors:
    - path is empty: the path value is empty.
    - path is absolute: the path starts with / or a
      drive letter like C:.
    - path contains backslash: the path contains \ characters.
    - directory traversal: the path contains .. components
      after normalization.
    - resolves outside root: after resolving symlinks, the
      path is outside the project root.

function ToCfsPath(os_path) -> CfsPath
  errors:
    - resolves outside root: the path is not within the
      project root.
```

### CfsPath

A path in the Code from Spec standard format:
- Forward slash (`/`) as separator, always.
- Relative to the project root.
- No `..` components, no drive letters, no leading `/`,
  no backslashes.

This is the only path format used in the framework's
public API — in frontmatter fields (`outputs`, `external`,
`input`), in logical names, and in tool parameters.

### OsPath

An absolute path in the operating system's native format:
- OS-specific separator (`/` on Unix, `\` on Windows).
- Always absolute.

This type is never exposed in the framework's public API.
It exists only inside the `os/` layer for interacting
with the filesystem.

### GetProjectRoot

Returns the project root as an `OsPath`. Determined from
the working directory of the process. Called once at
startup; the result is reused by `ResolvePath` and
`ToCfsPath`.

### ValidateCfsPath

Pure string validation — checks whether a value is a
valid `CfsPath` without touching the filesystem. Returns
true if the value is non-empty, uses only forward slashes,
is not absolute (no leading `/` or drive letter), and
contains no `..` components after normalization.

Use this for sanity checks on parameters received from
callers. Does not verify that the file exists or resolve
symlinks — use `ResolvePath` for that.

### ResolvePath

Validates a `CfsPath` and converts it to an absolute
`OsPath`. This is the single entry point for going from
framework paths to OS paths. If validation fails, no
conversion happens — an error is returned.

### ToCfsPath

Converts an absolute `OsPath` to a `CfsPath` relative to
the project root. Used internally by components that
receive paths from the OS (e.g. directory walking).

# Agent

Generate pseudocode for all functions.

## Implementation guidance

### ValidateCfsPath steps

1. Return false if the value is empty.
2. Return false if the value starts with `/` or a drive
   letter (e.g. `C:`).
3. Return false if the value contains `\`.
4. Normalize the path (resolve `.` and `..`).
5. Return false if any component is `..` after normalization.
6. Return true.

### ResolvePath steps

1. Reject empty paths.
2. Reject absolute paths (leading `/` or drive letter).
3. Reject paths containing backslash (`\`).
4. Normalize the path (resolve `.` and `..`).
5. Reject if any component is `..` after normalization.
6. Replace forward slashes with the OS separator.
7. Join with the project root to form an absolute path.
8. Resolve symlinks.
9. Verify the resolved path is within the project root.
10. Return the absolute OsPath.

### ToCfsPath steps

1. Resolve symlinks in the os_path.
2. Verify the resolved path is within the project root.
3. Compute the relative path from project root.
4. Replace OS separators with forward slashes.
5. Return the CfsPath.

### GetProjectRoot

Returns the working directory of the process as an
`OsPath`. `ResolvePath` and `ToCfsPath` use this
internally.

### Security

See `ROOT/functional/dependencies/owasp-path-traversal`
for the threat model. ResolvePath is the primary defense
against path traversal attacks.

## Contracts

- ResolvePath never creates or modifies files.
- ResolvePath never sanitizes — rejects invalid paths.
- ToCfsPath never creates or modifies files.
- Both functions resolve symlinks before checking
  containment.
