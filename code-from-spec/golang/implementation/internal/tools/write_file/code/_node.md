---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/implementation/internal/frontmatter
  - ROOT/golang/implementation/internal/logical_names
  - ROOT/golang/implementation/os/path_utils
input: ARTIFACT/functional/logic/mcp_tools/write_file(write_file)
outputs:
  - id: write_file
    path: internal/write_file/write_file.go
---

# ROOT/golang/implementation/internal/tools/write_file/code

Implementation of the write_file tool handler.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `mcp-go` SDK types for tool results.
- Call internal packages (`frontmatter`, `logicalnames`,
  `pathvalidation`) for their respective operations.
- Use `os.MkdirAll` for creating intermediate directories
  and `os.WriteFile` for writing the file.
- Use `filepath.ToSlash` to normalize paths to forward slashes.
- The package name should be `write_file`.

## Constraints

- The target argument must be a logical name that resolves to a
  node with `outputs`. Absent, empty, or invalid values cause
  the tool to report an error.
- Writes are limited to `outputs`.
- The validation against `outputs` is the security boundary of
  `write_file`. It must not be bypassable.
- Exactly one file is written per `write_file` call.
