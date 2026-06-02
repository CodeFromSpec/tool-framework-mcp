[//]: # (code-from-spec: ROOT/golang/interfaces/chain/resolver@DBcbnssMbfih0A7nqbH39Vg8WSU)

# Package `chainresolver`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
```

## Structs

```go
package chainresolver

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

type ChainItem struct {
	LogicalName string
	FilePath    pathutils.PathCfs
	Qualifier   string
}

type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	External     []*frontmatter.FrontmatterExternal
	Target       *ChainItem
	Input        *ChainItem
}
```

## Error Sentinels

```go
package chainresolver

import "errors"

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")
```

## Functions

```go
package chainresolver

// ChainResolve returns the chain for a target logical name — the ordered
// list of positions that a downstream tool needs to assemble context for
// artifact generation or to compute the chain hash.
//
// Chain assembly order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — entries from the target's depends_on, sorted
//     alphabetically by file path then by qualifier, each with its
//     resolved file path and an optional qualifier.
//  3. External — files from the target's external, sorted alphabetically
//     by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
func ChainResolve(targetLogicalName string) (*Chain, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		log.Fatal(err)
	}

	for _, ancestor := range chain.Ancestors {
		fmt.Println("Ancestor:", ancestor.LogicalName, ancestor.FilePath.Value)
	}

	for _, dep := range chain.Dependencies {
		fmt.Println("Dependency:", dep.LogicalName, dep.FilePath.Value, dep.Qualifier)
	}

	for _, ext := range chain.External {
		fmt.Println("External:", ext.Path)
	}

	fmt.Println("Target:", chain.Target.LogicalName, chain.Target.FilePath.Value)

	if chain.Input != nil {
		fmt.Println("Input:", chain.Input.LogicalName, chain.Input.FilePath.Value)
	}
}
```
