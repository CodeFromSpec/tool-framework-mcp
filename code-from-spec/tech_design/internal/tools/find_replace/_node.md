# ROOT/tech_design/internal/tools/find_replace

Implements the `find_replace` tool handler: finds a unique
string in an existing file and replaces it, after validating
the target path against the node's `implements` list and the
project root.

# Public

## Package

`package find_replace`

## Target node

The target node is identified by its logical name — either a leaf
spec node (`ROOT/...`) or a test node (`TEST/...`). Examples:
`ROOT/payments/fees/calculation`,
`TEST/payments/fees/calculation`.

## Interface

### Tool definition

Name: `find_replace`
Description: `"Replace a specific string in an existing source file. The old_string must appear exactly once. The path must be one of the files declared in the node's implements list. The file must already exist."`

Input parameters:

| Name | Type | Required | Description |
|---|---|---|---|
| `logical_name` | string | yes | Logical name of the node whose implements list authorizes the write. |
| `path` | string | yes | Relative file path from project root. |
| `old_string` | string | yes | Exact string to find in the file. Must match exactly once. |
| `new_string` | string | yes | Replacement string. |

### FindReplaceArgs type

```go
type FindReplaceArgs struct {
    LogicalName string `json:"logical_name" jsonschema:"Logical name of the node whose implements list authorizes the write."`
    Path        string `json:"path" jsonschema:"Relative file path from project root."`
    OldString   string `json:"old_string" jsonschema:"Exact string to find in the file. Must match exactly once."`
    NewString   string `json:"new_string" jsonschema:"Replacement string."`
}
```

### Handler

```go
func HandleFindReplace(
    ctx context.Context,
    req *mcp.CallToolRequest,
    args FindReplaceArgs,
) (*mcp.CallToolResult, any, error)
```
