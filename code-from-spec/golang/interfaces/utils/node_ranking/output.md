[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@A5P3QeZ1fkTeC2hNYAaKXZtGTY8)

# Package `noderanking`

```go
package noderanking
```

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"`

## Error Sentinels

```go
package noderanking

import "errors"

var ErrUnresolvableReference = errors.New("unresolvable reference")
```

## Struct Definitions

```go
package noderanking

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"

// NodeRankInput pairs a logical name with its parsed frontmatter, providing
// the dependency information needed to compute a topological rank.
type NodeRankInput struct {
	LogicalName string
	Frontmatter frontmatter.Frontmatter
}

// NodeRankEntry associates a logical name with its computed rank. Nodes with
// no dependencies have rank 0; nodes that depend on others have a rank one
// greater than the highest rank among their dependencies.
type NodeRankEntry struct {
	LogicalName string
	Rank        int
}
```

## Function Signatures

```go
package noderanking

// NodeRankCompute computes a topological rank for each node and artifact in
// the input set. Returns all ranked entries and a list of logical names
// involved in dependency cycles (empty when no cycles exist).
//
// Returns ErrUnresolvableReference if a depends_on or input target cannot be
// resolved within the known set of nodes.
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
			Frontmatter: frontmatter.Frontmatter{},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a"},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		log.Fatal(err)
	}

	if len(cycles) > 0 {
		fmt.Println("cycles detected:", cycles)
	}

	for _, r := range ranked {
		fmt.Printf("%s -> rank %d\n", r.LogicalName, r.Rank)
	}
}
```
