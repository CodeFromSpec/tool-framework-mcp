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
  `ErrWalkError`, `ErrSymlinkNotAllowed`) — declared
  in `errors.go`.

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
   every entry.

   For each entry encountered during the walk:
     If the walk callback receives a filesystem error
     (the `err` parameter is non-nil), raise ErrWalkError.
     If the entry is a directory, skip it (continue
     traversal but do not add).
     If the entry is a symlink, raise
     ErrSymlinkNotAllowed. Propagate this error directly
     — do not wrap it in ErrWalkError.
     If the entry is a regular file:
       Call OsPathToCfs(OsPath(entry_os_path)) to
       convert it to a CfsPath. If it raises any error,
       propagate it directly — do not wrap it in
       ErrWalkError.
       Append the resulting CfsPath to results.

   If the walk returns an error that was already raised
   by the callback (ErrSymlinkNotAllowed, or a
   propagated error from OsPathToCfs), return it as-is.
   Only wrap in ErrWalkError if the error originates
   from the walk mechanism itself.

5. Sort results alphabetically by their string value.

6. Return results.

## Go-specific guidance

- Use `filepath.WalkDir` for recursive directory
  traversal.
- Use CfsPathToOs and OsPathToCfs from this package
  for path conversions.
- Read-only — never create or modify files on disk.
