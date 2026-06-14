[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@obbyQHUUIVI8toSBeS7W1vCFixE)

# Package `noderanking`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking`

---

## Structs

```go
package noderanking

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"

// NodeRankInput represents a discovered node with its parsed frontmatter,
// used as input to NodeRankCompute.
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry holds a logical name and its computed rank.
type NodeRankEntry struct {
	LogicalName string
	Rank        int
}
```

---

## Error Sentinels

```go
package noderanking

import "errors"

var ErrUnresolvableReference = errors.New("a depends_on or input target cannot be resolved")
```

---

## Functions

```go
package noderanking

// NodeRankCompute takes the full set of discovered nodes with their parsed
// frontmatter. It returns ranked entries (nodes and artifacts) sorted by
// dependency order, and a list of logical names involved in cycles.
// Returns ErrUnresolvableReference if a depends_on or input target cannot
// be resolved.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

func main() {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "SPEC/payments",
			Frontmatter: &frontmatter.Frontmatter{},
		},
		{
			LogicalName: "SPEC/payments/fees",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"SPEC/payments"},
				Output:    "ARTIFACT/payments/fees",
			},
		},
		{
			LogicalName: "SPEC/reports/summary",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"SPEC/payments/fees"},
				Input:     "EXTERNAL/data/transactions.csv",
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		if errors.Is(err, noderanking.ErrUnresolvableReference) {
			log.Fatal("unresolvable reference in dependency graph:", err)
		}
		log.Fatal(err)
	}

	if len(cycles) > 0 {
		fmt.Println("Cycles detected:", cycles)
	}

	for _, entry := range ranked {
		fmt.Printf("rank %d: %s\n", entry.Rank, entry.LogicalName)
	}
}
```
