---
depends_on:
  - SPEC/golang/implementation/os/path_utils
output: internal/listfiles/listfiles.go
---

# SPEC/golang/implementation/os/list_files

Recursively lists all files under a directory.

# Public

## Package

`package listfiles`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/listfiles"`

## Interface

```go
func ListFiles(cfsPath pathutils.PathCfs) ([]pathutils.PathCfs, error)
```

Returns all files (not directories) found recursively
under the given directory. Results are `pathutils.PathCfs`
values, sorted alphabetically. If the directory exists
but contains no files, returns an empty list.

### Errors

- `ErrDirectoryNotFound`: the directory does not exist.
- `ErrWalkError`: a filesystem error occurred while
  traversing.
- Propagated errors from `pathutils` package.

# Agent

Implement the `listfiles` package, including its
interface.

## Logic

1. Call pathutils.PathCfsToOs(cfs_path) to get an OS
   path. If it raises any error, propagate it to the
   caller. Assign result to os_path.

2. Check that os_path points to an existing directory.
   If the directory does not exist, raise error
   "DirectoryNotFound".

3. Initialize an empty list, results.

4. Walk the directory at os_path recursively, visiting
   every entry. If the walk itself raises a filesystem
   error, raise error "WalkError".

   For each entry encountered during the walk:
     If the entry is a directory, skip it (continue
     traversal but do not add).
     If the entry is a file:
       Call PathOsToCfs(entry_os_path) to convert it
       to a PathCfs. If it raises any error, propagate
       it to the caller.
       Append the resulting PathCfs to results.

5. Sort results alphabetically by their value field.

6. Return results.

## Go-specific guidance

- Use `filepath.WalkDir` for recursive directory
  traversal.
- Use the `pathutils` package for `PathCfsToOs` and
  `PathOsToCfs` conversions.
- Read-only — never create or modify files on disk.
