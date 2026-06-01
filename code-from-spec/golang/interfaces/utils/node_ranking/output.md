[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@Y3p4zkYX29XRJliusCV9SLViJbc)

# Package `noderanking`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
```

Package `noderanking` computes a topological rank order for discovered spec nodes based on their frontmatter dependencies.

---

## Structs

```go
package noderanking

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"

// NodeRankInput holds a discovered node's logical name and its parsed
// frontmatter, used as input to NodeRankCompute.
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry holds a node or artifact logical name and its computed
// rank in the dependency order.
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

// ErrUnresolvableReference is returned when a depends_on or input
// target cannot be resolved to any known node.
var ErrUnresolvableReference = errors.New("unresolvable reference")
```

---

## Functions

```go
package noderanking

// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and returns a ranked list of entries (nodes and
// artifacts) along with a list of logical names involved in dependency
// cycles. The cycles list is empty when no cycles are detected.
//
// Errors:
//   - ErrUnresolvableReference: a depends_on or input target cannot
//     be resolved to any known node.
func NodeRankCompute(entries []*NodeRankInput) ([]*NodeRankEntry, []string, error)
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
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		if errors.Is(err, noderanking.ErrUnresolvableReference) {
			log.Fatal("a depends_on or input target could not be resolved")
		}
		log.Fatalf("rank compute failed: %v", err)
	}

	if len(cycles) > 0 {
		fmt.Println("cycles detected:", cycles)
	}

	for _, entry := range ranked {
		fmt.Printf("rank %d: %s\n", entry.Rank, entry.LogicalName)
	}
}
```
