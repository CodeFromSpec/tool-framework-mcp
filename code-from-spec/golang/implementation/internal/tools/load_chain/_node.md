# ROOT/golang/implementation/internal/tools/load_chain

Implements the `load_chain` tool handler: validates the
logical name, loads the spec chain, and returns the chain
content as a single MCP text response.

# Public

## Package

`package load_chain`

## Target node

The target node is identified by its logical name — a leaf
spec node (`ROOT/...`). Example:
`ROOT/payments/fees/calculation`.

## Interface

### Tool definition

Name: `load_chain`
Description: `"Load the spec chain context for a given logical name. Returns all relevant spec files concatenated in a single response."`

Input parameters:

| Name | Type | Required | Description |
|---|---|---|---|
| `logical_name` | string | yes | Logical name of the node to generate code for. |

### LoadChainArgs type

```go
type LoadChainArgs struct {
    LogicalName string `json:"logical_name" jsonschema:"Logical name of the node to generate code for."`
}
```

### Handler

```go
func HandleLoadChain(
    ctx context.Context,
    req *mcp.CallToolRequest,
    args LoadChainArgs,
) (*mcp.CallToolResult, any, error)
```

### Chain output format

The chain is returned as a single MCP text response containing
three concatenated text items, separated by blank lines:

1. **Hash** — a content hash of the chain for staleness
   detection.
2. **Context** — the assembled chain content. Each chain item
   is the spec node content (public section body for ancestors
   and dependencies, public + agent sections for the target).
   External files are included as raw file content.
3. **Input** — the target node's `input` frontmatter field.
   Empty string if no input is declared.

No heredoc delimiters, no `node:` or `path:` headers. The
chain content is plain concatenated text.

# Decisions

### load_chain returns everything in one call

Loading the chain file-by-file via separate tool calls would
accumulate context in the conversation history, increasing token
cost with each roundtrip. A single call returns the entire chain,
minimizing roundtrip overhead.
