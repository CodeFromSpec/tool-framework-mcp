[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@mrT-2tkesK7QPwuvQGNeRR7Szs4)

# Package `noderanking`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
```

Takes the full set of discovered nodes with their parsed frontmatter. Returns ranked entries (nodes and artifacts) and a list of logical names involved in cycles (empty if no cycles).

---

## Structs

```go
package noderanking

// NodeRankInput holds a discovered node's logical name and its parsed frontmatter,
// used as input to NodeRankCompute.
type NodeRankInput struct {
	// LogicalName is the ROOT/ logical name of the node.
	LogicalName string

	// Frontmatter is the parsed frontmatter of the node's spec file.
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry holds a node or artifact's logical name and its computed rank.
type NodeRankEntry struct {
	// LogicalName is the ROOT/ or ARTIFACT/ logical name.
	LogicalName string

	// Rank is the computed topological rank (lower rank = fewer dependencies).
	Rank int
}
```

---

## Error Sentinels

```go
package noderanking

import "errors"

// ErrUnresolvableReference is returned when a depends_on or input target
// cannot be resolved to a known node.
var ErrUnresolvableReference = errors.New("unresolvable reference")
```

---

## Functions

```go
package noderanking

// NodeRankCompute takes the full set of discovered nodes with their parsed
// frontmatter and computes a topological ranking for all nodes and artifacts.
//
// It returns:
//   - ranked: a list of NodeRankEntry values, one per node and artifact,
//     ordered by ascending rank.
//   - cycles: a list of logical names that are part of dependency cycles.
//     Empty if no cycles are detected.
//
// Errors:
//   - ErrUnresolvableReference: a depends_on or input target cannot be resolved.
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
				DependsOn: []string{"ROOT/b"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		log.Fatalf("NodeRankCompute: %v", err)
	}

	if len(cycles) > 0 {
		fmt.Println("cycles detected:", cycles)
	}

	for _, entry := range ranked {
		fmt.Printf("logical_name: %s, rank: %d\n", entry.LogicalName, entry.Rank)
	}
	// Output:
	// logical_name: ROOT/b, rank: 0
	// logical_name: ROOT/a, rank: 1
}
```
