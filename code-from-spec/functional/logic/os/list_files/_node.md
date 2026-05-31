---
depends_on:
  - ROOT/functional/logic/os/path_utils(interface)
outputs:
  - id: list_files
    path: code-from-spec/functional/logic/os/list_files/output.md
---

# ROOT/functional/logic/os/list_files

Recursively lists all files under a directory.

# Public

## Interface

```
function ListFiles(cfs_path: PathCfs) -> list of PathCfs
  errors:
    - DirectoryNotFound: the directory does not exist.
    - WalkError: a filesystem error occurred while
      traversing.
    - (PathUtils.*): propagated from PathCfsToOs.
    - (PathUtils.*): propagated from PathOsToCfs.
```

Returns all files (not directories) found recursively
under the given directory. Results are `PathCfs` values,
sorted alphabetically. If the directory exists but
contains no files, returns an empty list.

# Agent

Generate pseudocode for the ListFiles function.

## Implementation guidance

- Convert `cfs_path` to an OS path using `PathCfsToOs`.
- Walk the directory recursively.
- For each file found, convert the OS path back to a
  `PathCfs` using `PathOsToCfs`.
- Only include files in the result — directories are
  traversed but not themselves included.
- Sort the result alphabetically by value.
