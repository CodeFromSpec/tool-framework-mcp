---
depends_on:
  - ROOT/functional/logic/os/path_utils
outputs:
  - id: list_files
    path: code-from-spec/functional/logic/os/list_files/output.md
---

# ROOT/functional/logic/os/list_files

Recursively lists all files under a directory.

Review status: pending

# Public

## Interface

```
function ListFiles(cfs_path) -> list of CfsPath
  errors:
    - directory not found: the directory does not exist.
    - walk error: a filesystem error occurred while
      traversing.
```

Returns all files (not directories) found recursively
under the given directory. Results are `CfsPath` values,
sorted alphabetically.

# Agent

Generate pseudocode for the ListFiles function.

## Implementation guidance

- Convert `cfs_path` to an OS path using `path_utils`.
- Walk the directory recursively.
- For each file found, convert the OS path back to a
  `CfsPath` using `ToCfsPath`.
- Skip directories — only include files.
- Sort the result alphabetically by value.
