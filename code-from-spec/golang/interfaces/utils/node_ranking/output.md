[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@m7-WUpO0HYHgO5En44CjBqLl5tQ)

# Package `noderanking`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
```

## Structs

```go
package noderanking

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"

type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

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
// frontmatter. Returns ranked entries (nodes and artifacts) and a list of
// logical names involved in cycles (empty if no cycles).
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error)
```

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
			Frontmatter: &frontmatter.Frontmatter{},
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
		fmt.Printf("Node: %s, Rank: %d\n", entry.LogicalName, entry.Rank)
	}
}
```
