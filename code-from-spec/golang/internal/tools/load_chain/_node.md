# ROOT/golang/internal/tools/load_chain

Implements the `load_chain` tool handler: validates the
logical name, loads the spec chain, and returns the chain
content as a single MCP text response.

# Public

## Package

`package load_chain`

## Target node

The target node is identified by its logical name — either a leaf
spec node (`ROOT/...`) or a test node (`TEST/...`). Examples:
`ROOT/payments/fees/calculation`,
`TEST/payments/fees/calculation`.

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

The chain is serialized as a sequence of file sections using
heredoc-style delimiters with a UUID generated once per call
to avoid collisions with file content.

Opening delimiter: `<<<FILE_<uuid>>>`
Closing delimiter: `<<<END_FILE_<uuid>>>`

The same UUID is used for all files in the chain. Each section
includes `node:` and `path:` headers between the opening
delimiter and the file content, separated by a blank line.
Code files include only `path:`.

```
<<<FILE_550e8400-e29b-41d4-a716-446655440000>>>
node: ROOT
path: code-from-spec/_node.md

<Public section body — no # Public heading>
<<<END_FILE_550e8400-e29b-41d4-a716-446655440000>>>

<<<FILE_550e8400-e29b-41d4-a716-446655440000>>>
node: ROOT/payments/fees/calculation
path: code-from-spec/payments/fees/calculation/_node.md

<target content with reduced frontmatter>
<<<END_FILE_550e8400-e29b-41d4-a716-446655440000>>>

<<<FILE_550e8400-e29b-41d4-a716-446655440000>>>
node: ROOT/architecture/backend
path: code-from-spec/architecture/backend/_node.md

<Public section body — no # Public heading>
<<<END_FILE_550e8400-e29b-41d4-a716-446655440000>>>

<<<FILE_550e8400-e29b-41d4-a716-446655440000>>>
path: internal/payments/fees/calculation.go

<existing source file content>
<<<END_FILE_550e8400-e29b-41d4-a716-446655440000>>>
```

# Decisions

### load_chain returns everything in one call

Loading the chain file-by-file via separate tool calls would
accumulate context in the conversation history, increasing token
cost with each roundtrip. A single call returns the entire chain,
minimizing roundtrip overhead.
