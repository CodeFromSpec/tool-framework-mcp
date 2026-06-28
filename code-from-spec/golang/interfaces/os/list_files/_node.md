---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/os/list_files/output.md
---

# SPEC/golang/interfaces/os/list_files

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

Generate an interface specification document listing
the package, import path, and function signatures.
