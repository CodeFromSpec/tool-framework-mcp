---
depends_on:
  - ARTIFACT/golang/interfaces/os/list_files(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/logic/os/list_files(list_files)
outputs:
  - id: listfiles
    path: internal/listfiles/listfiles.go
---

# ROOT/golang/implementation/os/list_files

# Agent

Implement the `listfiles` package, including its interface.

## Go-specific guidance

- Use `filepath.WalkDir` for recursive directory traversal.
- Use the `pathutils` package for `PathCfsToOs` and
  `PathOsToCfs` conversions.
- Read-only — never create or modify files on disk.
