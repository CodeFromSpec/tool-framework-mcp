---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
output: code-from-spec/golang/interfaces/mcp_tools/write_file/output.md
---

# SPEC/golang/interfaces/mcp_tools/write_file

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

Generate an interface specification document listing
the package, import path, and function signatures.
