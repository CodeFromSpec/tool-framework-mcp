---
depends_on:
  - ROOT/functional/logic/os/path_utils(interface)
outputs:
  - id: file_writer
    path: code-from-spec/functional/logic/os/file_writer/output.md
---

# ROOT/functional/logic/os/file_writer

Writes content to a file on disk.

# Public

## Interface

```
function FileWrite(cfs_path: PathCfs, content: string)
  errors:
    - (path errors): propagated from PathCfsToOs if the
      path is invalid.
    - cannot create directory: an intermediate directory
      cannot be created.
    - cannot write file: the file cannot be written.
```

`FileWrite` writes `content` to the file at `cfs_path`
as UTF-8 encoded text. If the file exists, it is
overwritten. If it does not exist, it is created.
Intermediate directories are created as needed.

Content is written exactly as received — no
normalization of line endings or other transformations.

The path is validated before writing — if validation
fails, no file or directory is created.

# Agent

Generate pseudocode for the FileWrite function.

## Implementation guidance

- Convert `cfs_path` to an OS path using `PathCfsToOs`.
  If it raises an error, propagate it.
- Create intermediate directories if they do not exist.
- Write the content as UTF-8 encoded text.
- Overwrite existing files without warning.
