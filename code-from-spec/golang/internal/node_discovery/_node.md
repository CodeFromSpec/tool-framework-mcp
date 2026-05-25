# ROOT/golang/internal/node_discovery

Walks the filesystem to discover all spec nodes in the
spec tree.

# Public

## Package

`package nodediscovery`

## Interface

```go
type DiscoveredNode struct {
    LogicalName string
    FilePath    string
}

var (
    ErrDirNotFound  = errors.New("directory not found")
    ErrWalk         = errors.New("walk error")
    ErrNoNodesFound = errors.New("no nodes found")
)

func DiscoverNodes() ([]DiscoveredNode, error)
```

`DiscoverNodes` walks `code-from-spec/` relative to the
working directory and returns every `_node.md` file found,
with its logical name derived via `logicalnames` reverse
resolution.

The returned slice is sorted alphabetically by logical name.

### Error handling

All errors wrap a sentinel so callers can use `errors.Is()`:

| Sentinel | Returned when |
|---|---|
| `ErrDirNotFound` | `code-from-spec/` does not exist. |
| `ErrWalk` | Filesystem error while traversing. |
| `ErrNoNodesFound` | No `_node.md` files found. |
