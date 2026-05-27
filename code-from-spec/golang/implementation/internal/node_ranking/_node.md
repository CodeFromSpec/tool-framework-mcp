# ROOT/golang/implementation/internal/node_ranking

Detects circular references in the spec tree using
iterative ranking.

# Public

## Package

`package noderanking`

## Interface

```go
type RankedEntry struct {
    LogicalName string
    Rank        int
}

var ErrUnresolvableRef = errors.New("unresolvable reference")

func DetectCycles(nodes []nodediscovery.DiscoveredNode) ([]RankedEntry, []string, error)
```

`DetectCycles` takes the full set of discovered nodes with
their parsed frontmatter. Returns the ranked entries and a
slice of logical names involved in cycles (empty if no
cycles exist).

### Error handling

| Sentinel | Returned when |
|---|---|
| `ErrUnresolvableRef` | A `depends_on` or `input` target cannot be resolved. |
