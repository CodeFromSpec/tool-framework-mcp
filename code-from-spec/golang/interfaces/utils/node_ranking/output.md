[//]: # (code-from-spec: ROOT/golang/interfaces/utils/node_ranking@LZKD3C41cIB45Q-Gp6iZtKsVD04)

# Interface: `noderanking`

## Package

```go
package noderanking
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/noderanking"
```

---

## Struct Definitions

```go
// NodeRankInput represents a single discovered node and its parsed
// frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry represents a single ranked node or artifact, identified
// by its logical name and assigned a numeric rank.
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
	// target cannot be resolved to any known node in the input set.
	ErrUnresolvableReference = errors.New("unresolvable reference")
)
```

---

## Functions

```go
// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and computes a topological ranking. It returns
// the ranked entries (nodes and artifacts) and a list of logical names
// involved in dependency cycles (empty if no cycles exist).
//
// Possible errors:
//   - ErrUnresolvableReference: a depends_on or input target cannot be
//     resolved to any known node in the input set.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/noderanking"
)

func main() {
	dep := "ROOT/x/base"
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT/x/base",
			Frontmatter: &frontmatter.Frontmatter{},
		},
		{
			LogicalName: "ROOT/x/derived",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{&dep},
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
		fmt.Printf("LogicalName=%s Rank=%d\n", entry.LogicalName, entry.Rank)
	}
}
```
