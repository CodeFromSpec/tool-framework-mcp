# ROOT/golang/implementation/internal/tools/write_file

Implements the `write_file` tool handler: resolves the node's
frontmatter from the provided logical name, validates the
target path against the `outputs` list and the project root,
then writes the file to disk.

# Public

## Package

`package write_file`

## Target node

The target node is identified by its logical name — a leaf
spec node (`ROOT/...`). Example:
`ROOT/payments/fees/calculation`.

## Interface

### Tool definition

Name: `write_file`
Description: `"Write a generated source file to disk. The path must be one of the files declared in the node's outputs list. Overwrites existing content."`

Input parameters:

| Name | Type | Required | Description |
|---|---|---|---|
| `logical_name` | string | yes | Logical name of the node whose outputs list authorizes the write. |
| `path` | string | yes | Relative file path from project root. |
| `content` | string | yes | Complete file content to write. |

### WriteFileArgs type

```go
type WriteFileArgs struct {
    LogicalName string `json:"logical_name" jsonschema:"Logical name of the node whose outputs list authorizes the write."`
    Path        string `json:"path" jsonschema:"Relative file path from project root."`
    Content     string `json:"content" jsonschema:"Complete file content to write."`
}
```

### Handler

```go
func HandleWriteFile(
    ctx context.Context,
    req *mcp.CallToolRequest,
    args WriteFileArgs,
) (*mcp.CallToolResult, any, error)
```

# Decisions

### write_file validates against outputs

The target node's `outputs` field is the authoritative list of
files this tool may produce. Validating every write against it
prevents the subagent from writing to paths outside the declared
scope, whether by mistake or hallucination.
