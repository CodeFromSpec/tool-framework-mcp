---
outputs:
  - id: write_file_test
    path: internal/write_file/write_file_test.go
---

# ROOT/golang/internal/tools/write_file/tests

Tests for the write_file tool handler.

# Agent

## Context

Each test uses `t.TempDir()` as the project root and working
directory. A spec tree is created with the necessary frontmatter
containing an `Outputs` list. The handler is called with
`WriteFileArgs` including the `LogicalName` of the node.

## Happy Path

### Writes file successfully

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: output/file.go`. Call the
handler with `LogicalName: "ROOT/a"`,
`Path: "output/file.go"`, and `Content: "package main"`.

Expect: success result with text `"wrote output/file.go"`.
Verify the file exists on disk with the correct content.

### Creates intermediate directories

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: deep/nested/dir/file.go`.
Call the handler with `Path: "deep/nested/dir/file.go"`.

Expect: success. Directories created automatically.

### Overwrites existing file

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: output/file.go`. Write an
initial file at that path. Call the handler with new content.

Expect: success. File content replaced.

### Path with backslashes is normalized (Windows only)

Skip this test on non-Windows platforms — backslash is a
valid filename character on Linux/macOS, not a separator.

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: output/file.go`. Call the
handler with `LogicalName: "ROOT/a"`,
`Path: "output\\file.go"`, and `Content: "package main"`.

Expect: success result with text `"wrote output/file.go"`.
The backslash path matches the forward-slash outputs entry
after normalization.

## Failure Cases

### Invalid logical name prefix

Call the handler with `LogicalName: "ROOT/external/something"`.

Expect: tool error.

### Nonexistent logical name

Call the handler with `LogicalName: "ROOT/nonexistent"`.
Do not create the corresponding spec file.

Expect: tool error.

### Path not in outputs

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: allowed/file.go`. Call the
handler with `Path: "other/file.go"`.

Expect: tool error containing `"path not allowed"` and
listing the allowed paths.

### Path traversal attempt

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: ../../etc/passwd`. Call
the handler with `Path: "../../etc/passwd"`.

Expect: tool error from `ValidatePath`.

### Empty path

Create a spec tree with `ROOT/a` having
`outputs:\n  - id: file\n    path: some/file.go`. Call the
handler with `Path: ""`.

Expect: tool error containing `"path is empty"`.

### Symlink escaping project root

Create a symlink inside the temp dir pointing outside it.
Create a spec tree with the symlink path in `outputs`.
Call the handler with that path.

Expect: tool error containing `"resolves outside project root"`.
