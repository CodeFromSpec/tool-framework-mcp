---
depends_on:
  - ROOT/functional/logic/os/path_utils
outputs:
  - id: file_writer
    path: code-from-spec/functional/logic/os/file_writer/output.md
---

# ROOT/functional/logic/os/file_writer

Writes content to a file on disk. Creates intermediate
directories if they do not exist. Validates the path
before writing.

Review status: pending

# Public

## Interface

```
function WriteFile(cfs_path, content) -> void
  errors:
    - path validation failed: the path does not pass
      validation (empty, absolute, traversal, or outside
      project root).
    - cannot create directory: an intermediate directory
      cannot be created.
    - cannot write file: the file cannot be written.
```

`WriteFile` writes `content` to the file at `cfs_path`.
If the file exists, it is overwritten. If it does not
exist, it is created. Intermediate directories are
created as needed.

The path is validated before writing — if validation
fails, no file or directory is created.

# Agent

Generate pseudocode for the WriteFile function.

## Implementation guidance

- Convert `cfs_path` to an OS path internally using
  the path conversion from `path_utils`.
- Validate the path using `ValidatePath` before any
  filesystem operation.
- Create intermediate directories if they do not exist.
- Write the full content atomically where possible.
- Overwrite existing files without warning.
