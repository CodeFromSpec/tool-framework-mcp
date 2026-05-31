[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@pJDRGsHaqmdAQXbTtZt5bip7Uro)

# Package `noderanking`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
```

Package `noderanking` computes a topological ranking of spec nodes and artifacts based on their dependency declarations. It detects cycles in the dependency graph and returns the ordered list of nodes along with any cycle members.

---

## Structs

```go
package noderanking

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"

// NodeRankInput represents a discovered node with its parsed
// frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
	// LogicalName is the ROOT/ logical name of the node.
	LogicalName string

	// Frontmatter holds the parsed frontmatter for the node,
	// including its dependency declarations.
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry represents a node or artifact in the ranked output.
type NodeRankEntry struct {
	// LogicalName is the ROOT/ or ARTIFACT/ logical name.
	LogicalName string

	// Rank is the computed topological rank of this entry.
	// Lower values come earlier in the dependency order.
	Rank int
}
```

---

## Error Sentinels

```go
package noderanking

import "errors"

// ErrUnresolvableReference is returned when a depends_on or input
// target in the frontmatter cannot be resolved to a known node.
var ErrUnresolvableReference = errors.New("unresolvable reference")
```

---

## Functions

```go
package noderanking

// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and returns a topologically ranked list of entries
// (nodes and artifacts) along with the logical names of any nodes
// involved in dependency cycles.
//
// The returned ranked slice contains one entry per node/artifact in
// dependency order (lowest rank first). The cycles slice contains the
// logical names of all nodes that participate in a cycle; it is empty
// when no cycles are detected.
//
// Errors:
//   - ErrUnresolvableReference: a depends_on or input target in the
//     frontmatter cannot be resolved to a known node.
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
	// Build the input set from discovered nodes and their parsed
	// frontmatter. In practice these come from the spec tree scan.
	inputs := []*noderanking.NodeRankInput{
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

	// Compute the topological ranking.
	ranked, cycles, err := noderanking.NodeRankCompute(inputs)
	if err != nil {
		if errors.Is(err, noderanking.ErrUnresolvableReference) {
			log.Fatal("a depends_on or input target could not be resolved")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	// Report any detected cycles.
	if len(cycles) > 0 {
		fmt.Println("dependency cycles detected:")
		for _, name := range cycles {
			fmt.Println(" ", name)
		}
	}

	// Print the ranked entries in dependency order.
	for _, entry := range ranked {
		fmt.Printf("rank=%d  %s\n", entry.Rank, entry.LogicalName)
	}
}
```
