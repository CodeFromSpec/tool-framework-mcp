[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@f24gZMybXZxk2rIPeYt_AOK5Rxs)

# Interface: `noderanking`

**Package:** `package noderanking`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"`

---

## Structs

```go
// NodeRankInput represents a discovered node with its logical name and
// parsed frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
    LogicalName string
    Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry represents a node or artifact with its computed
// topological rank.
type NodeRankEntry struct {
    LogicalName string
    Rank        int
}
```

---

## Error Sentinels

```go
var (
    // ErrUnresolvableReference is returned when a depends_on or input
    // target cannot be resolved to a known node.
    ErrUnresolvableReference = errors.New("unresolvable reference")
)
```

---

## Functions

```go
// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and computes a topological rank for each node
// and artifact. It returns the ranked entries and a list of logical
// names involved in dependency cycles (empty if no cycles exist).
//
// Returns ErrUnresolvableReference if a depends_on or input target
// cannot be resolved to any entry in the provided set.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error)
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

func main() {
    entries := []*noderanking.NodeRankInput{
        {
            LogicalName: "ROOT/a",
            Frontmatter: &frontmatter.Frontmatter{
                DependsOn: []string{},
            },
        },
        {
            LogicalName: "ROOT/b",
            Frontmatter: &frontmatter.Frontmatter{
                DependsOn: []string{"ROOT/a"},
            },
        },
        {
            LogicalName: "ROOT/c",
            Frontmatter: &frontmatter.Frontmatter{
                DependsOn: []string{"ROOT/a", "ROOT/b"},
            },
        },
    }

    ranked, cycles, err := noderanking.NodeRankCompute(entries)
    if err != nil {
        log.Fatalf("failed to compute node ranks: %v", err)
    }

    if len(cycles) > 0 {
        fmt.Println("cycles detected:", cycles)
    }

    for _, entry := range ranked {
        fmt.Printf("node: %s, rank: %d\n", entry.LogicalName, entry.Rank)
    }
    // Output (order may vary within same rank):
    // node: ROOT/a, rank: 0
    // node: ROOT/b, rank: 1
    // node: ROOT/c, rank: 2
}
```
