# ROOT/golang/implementation/internal/tools/validate_specs

Implements the `validate_specs` tool handler: discovers all
spec nodes, validates format, detects circular references,
and checks artifact staleness across the entire spec tree.

# Public

## Package

`package validate_specs`

## Interface

### Tool definition

Name: `validate_specs`
Description: `"Validate the spec tree for format errors, circular references, and artifact staleness."`

No input parameters. Scans the entire spec tree.

### ValidateSpecsArgs type

```go
type ValidateSpecsArgs struct{}
```

### Handler

```go
func HandleValidateSpecs(
    ctx context.Context,
    req *mcp.CallToolRequest,
    args ValidateSpecsArgs,
) (*mcp.CallToolResult, any, error)
```
