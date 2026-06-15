[//]: # (code-from-spec: SPEC/golang/interfaces/utils/node_ranking@_GE-wze4pr2JQ06vZi26j_OuHPQ)

# Package `noderanking`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/noderanking`

## Types

```go
package noderanking

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"

// NodeRankInput holds a discovered node's logical name and its parsed
// frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry holds the logical name and computed rank of a node or artifact.
type NodeRankEntry struct {
	LogicalName string
	Rank        int
}
```

## Error Sentinels

```go
package noderanking

import "errors"

var ErrUnresolvableReference = errors.New("unresolvable reference")
```

## Functions

```go
package noderanking

// NodeRankCompute takes the full set of discovered nodes with their parsed
// frontmatter and computes a topological rank for each node and artifact.
// Returns the ranked entries and a list of logical names involved in cycles.
// Returns ErrUnresolvableReference if a depends_on or input target cannot
// be resolved.
func NodeRankCompute(entries []*NodeRankInput) ([]*NodeRankEntry, []string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/noderanking"
)

func main() {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "SPEC/payments/fees",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"SPEC/payments/core"},
				Output:    "ARTIFACT/payments/fees",
			},
		},
		{
			LogicalName: "SPEC/payments/core",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		log.Fatal(err)
	}

	if len(cycles) > 0 {
		fmt.Println("Cycles detected:", cycles)
	}

	for _, entry := range ranked {
		fmt.Printf("LogicalName: %s  Rank: %d\n", entry.LogicalName, entry.Rank)
	}
}
```
