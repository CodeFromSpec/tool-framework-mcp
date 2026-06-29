---
output: internal/oslayer/list_all_files.go
---

# SPEC/golang/implementation/oslayer/list_all_files

Recursively lists all files under a directory.

# Agent

Implement the function listed in the Ownership section
as a Go file in package `oslayer`.

## Ownership

This file declares and implements:
- Functions: `ListAllFiles`

The following exist in other files of this package and
can be used but must not be redeclared:
- Types: `CfsPath`, `OsPath` — declared in `path.go`.
- Functions: `CfsPathToOs`, `OsPathToCfs` — declared
  in `path.go`.
- Error sentinels (`ErrDirectoryNotFound`,
  `ErrWalkError`) — declared in `errors.go`.

To avoid name collisions with other files in this
package, all identifiers you declare beyond the ones
listed in the Ownership section (functions, variables,
types) must use the suffix `List`.

## Logic

1. Call CfsPathToOs(cfsPath) to get an OS path. If it
   raises any error, propagate it. Assign result to
   os_path.

2. Check that os_path points to an existing directory.
   If the directory does not exist, raise
   ErrDirectoryNotFound.

3. Initialize an empty list, results.

4. Walk the directory at os_path recursively, visiting
   every entry. If the walk itself raises a filesystem
   error, raise ErrWalkError.

   For each entry encountered during the walk:
     If the entry is a directory, skip it (continue
     traversal but do not add).
     If the entry is a file:
       Call OsPathToCfs(OsPath(entry_os_path)) to
       convert it to a CfsPath. If it raises any error,
       propagate it.
       Append the resulting CfsPath to results.

5. Sort results alphabetically by their string value.

6. Return results.

## Go-specific guidance

- Use `filepath.WalkDir` for recursive directory
  traversal.
- Use CfsPathToOs and OsPathToCfs from this package
  for path conversions.
- Read-only — never create or modify files on disk.
