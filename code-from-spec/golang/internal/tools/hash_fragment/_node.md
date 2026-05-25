# ROOT/golang/internal/tools/hash_fragment

Implements the `hash_fragment` tool handler: validates the
file path, reads the specified line range, and returns a
SHA-1 hash (base64url encoded) of the extracted content.

# Public

## Package

`package hash_fragment`

## Interface

### Tool definition

Name: `hash_fragment`
Description: `"Calculate the hash of a line range in a file, for use in external: fragment declarations."`

Input parameters:

| Name | Type | Required | Description |
|---|---|---|---|
| `path` | string | yes | File path relative to project root. |
| `lines` | string | yes | Line range (e.g., `"150-210"`). |

### HashFragmentArgs type

```go
type HashFragmentArgs struct {
    Path  string `json:"path" jsonschema:"File path relative to project root."`
    Lines string `json:"lines" jsonschema:"Line range (e.g., 150-210)."`
}
```

### Handler

```go
func HandleHashFragment(
    ctx context.Context,
    req *mcp.CallToolRequest,
    args HashFragmentArgs,
) (*mcp.CallToolResult, any, error)
```
