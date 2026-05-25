---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/internal/frontmatter
  - ROOT/golang/internal/logical_names
  - ROOT/golang/internal/pathvalidation
outputs:
  - id: write_file
    path: internal/write_file/write_file.go
---

# ROOT/golang/internal/tools/write_file/code

Implementation of the write_file tool handler.

# Agent

## Implementation

1. Validate that `args.LogicalName` starts with `ROOT/`
   (or equals `ROOT`). If not, return a tool error.
2. Call `logicalnames.PathFromLogicalName`. If it returns false, return a
   tool error: `"invalid logical name: <name>"`.
3. Call `ParseFrontmatter` on the resolved path. If it fails,
   return a tool error wrapping the underlying error.
4. Validate `Outputs` is not empty → tool error:
   `"node <name> has no outputs"`.
5. Normalize `args.Path` to forward slashes using
   `filepath.ToSlash`.
6. Call `ValidatePath` on the normalized path against the
   working directory. If it fails, return a tool error with
   the validation error and the list of valid `outputs`
   paths.
7. Check that the normalized path appears in the `Path` fields
   of the frontmatter's `Outputs` (exact string match). If
   not, return a tool error listing the valid paths.
8. Create any missing intermediate directories for the target
   path.
9. Write `args.Content` to the file, overwriting if it exists.
10. Return a success result with text `"wrote <path>"`.

### Error handling

- Invalid logical name → tool error with the name.
- Frontmatter parse failure → tool error wrapping the error.
- No outputs → tool error: `"node <name> has no outputs"`.
- Path validation failure → tool error with the violation and
  the list of allowed paths.
- Path not in outputs → tool error: `"path not allowed:
  <path>. allowed paths: <list>"`.
- Directory creation failure → tool error: `"failed to create
  directories for <path>: <err>"`.
- Write failure → tool error: `"failed to write <path>:
  <err>"`.

## Constraints

- The target argument must be a logical name that resolves to a
  node with `outputs`. Absent, empty, or invalid values cause
  the tool to report an error.
- Writes are limited to `outputs`.
- The validation against `outputs` is the security boundary of
  `write_file`. It must not be bypassable.
- Exactly one file is written per `write_file` call.
