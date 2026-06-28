---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
output: internal/mcpwritefile/mcpwritefile.go
---

# SPEC/golang/implementation/mcp_tools/write_file

Writes a generated source file to disk after validating
the path against the node's declared output.

# Public

## Package

`package mcpwritefile`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpwritefile"`

## Interface

```go
func MCPWriteFile(logicalName, path, content string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the node whose output authorizes the write. |
| `path` | yes | Relative file path from project root (forward slashes). |
| `content` | yes | Complete file content (UTF-8 text). |

### Output

A success message: `"wrote <path>"`.

### Errors

- `ErrNotASpecReference`: the logical name is not a
  SPEC/ reference.
- `ErrQualifierNotAllowed`: the logical name contains
  a parenthetical qualifier.
- `ErrUnreadableFrontmatter`: the node's frontmatter
  cannot be parsed.
- `ErrNoOutput`: target node has no output field.
- `ErrPathNotInOutput`: path is not declared in the
  node's output.
- Propagated errors from `logicalnames`, `pathutils`,
  `file` packages.

# Agent

Implement the write file tool as a Go package.

## Logic

1. If logical_name does not start with "SPEC/" and
   is not exactly "SPEC", return error
   "not a SPEC reference".

2. Call `LogicalNameParse(logical_name)`.
   If it fails, propagate the error.
   Let `ln` be the result.

3. If ln.Qualifier is not nil, return error
   "qualifier not allowed".

4. Call `FrontmatterParse(PathCfs{Value: ln.Path})`.
   If it fails, return error "unreadable frontmatter".
   Store the result as frontmatter.

5. If `frontmatter.output` is empty, return error
   "no output".

6. Call `PathValidateCfs` with path. If it fails,
   propagate the error.

7. If path does not exactly match `frontmatter.output`,
   return error "path not in output".

8. Construct a `PathCfs` record with value set to path.
   Call `FileOpen` with that PathCfs, mode "overwrite",
   and timeout 30000. If it fails, propagate the error.
   Store the result as handle.

9. Call `FileWrite` with handle and content. If it
   fails, call `FileClose` with handle, then propagate
   the error.

10. Call `FileClose` with handle.

11. Return "wrote <path>" where <path> is the path
    string.

## Go-specific guidance

- Use the `logicalnames` package for `LogicalNameParse`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `file` package for `FileOpen`, `FileWrite`,
  `FileClose`.
- The package name should be `mcpwritefile`.
- The function receives plain strings from the MCP
  transport layer. Construct `PathCfs` internally.
