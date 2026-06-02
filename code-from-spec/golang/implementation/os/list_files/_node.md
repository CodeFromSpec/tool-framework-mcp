---
depends_on:
  - ARTIFACT/golang/interfaces/os/list_files
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/os/list_files
output: internal/listfiles/listfiles.go
---

# ROOT/golang/implementation/os/list_files

# Agent

Implement the `listfiles` package, including its interface.

## Go-specific guidance

- Use `filepath.WalkDir` for recursive directory traversal.
- Use the `pathutils` package for `PathCfsToOs` and
  `PathOsToCfs` conversions.
- Read-only — never create or modify files on disk.
