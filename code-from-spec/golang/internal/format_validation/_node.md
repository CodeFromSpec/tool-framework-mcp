# ROOT/golang/internal/format_validation

Linter for spec nodes. Reads every node in the spec tree,
parses its frontmatter and body, and checks structural rules
defined by the framework. Reports all violations found.

# Public

## Package

`package formatvalidation`

## Interface

```go
type FormatError struct {
    Node   string
    Rule   string
    Detail string
}

var ErrUnreadableNode = errors.New("unreadable node")

func ValidateFormat(discoveredNodes []nodediscovery.DiscoveredNode) ([]FormatError, error)
```

`ValidateFormat` takes a list of discovered nodes (logical
name + file path) and validates each one against structural
rules. Returns a slice of format errors (empty if all nodes
are valid).

A node is classified as leaf or intermediate by checking
whether any other discovered node is a child of it.

### Error handling

| Sentinel | Returned when |
|---|---|
| `ErrUnreadableNode` | A node file cannot be read. |
