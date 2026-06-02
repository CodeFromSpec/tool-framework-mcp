[//]: # (code-from-spec: ROOT/golang/interfaces/chain/resolver@wdK7kizmQPUjbo9leC8yPZfs2Mg)

# Package `chainresolver`

```go
package chainresolver
```

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver`

## Types

```go
package chainresolver

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ChainItem represents a single node in the resolved chain.
type ChainItem struct {
	LogicalName string
	FilePath    pathutils.PathCfs
	Qualifier   string
}

// Chain holds the fully resolved chain for a target node.
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
var ErrUnresolvableArtifact  = errors.New("unresolvable artifact")
```

## Functions

```go
package chainresolver

// ChainResolve resolves the full chain for the given target logical name.
// It collects ancestors from root down to (but not including) the target,
// resolves depends_on entries as dependencies sorted alphabetically by file
// path then qualifier, collects external entries sorted alphabetically by
// path, identifies the target itself, and resolves the input artifact if
// present.
//
// Returns ErrUnreadableFrontmatter if any node's frontmatter cannot be
// parsed, ErrUnresolvableArtifact if an ARTIFACT/ reference cannot be
// resolved, or propagated errors from LogicalNames and Frontmatter packages.
func ChainResolve(target_logical_name string) (*Chain, error)
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

	fmt.Println("target:", chain.Target.LogicalName)

	for _, ancestor := range chain.Ancestors {
		fmt.Println("ancestor:", ancestor.LogicalName)
	}

	for _, dep := range chain.Dependencies {
		fmt.Println("dependency:", dep.LogicalName, "qualifier:", dep.Qualifier)
	}

	for _, ext := range chain.External {
		fmt.Println("external:", ext.Path)
	}

	if chain.Input != nil {
		fmt.Println("input:", chain.Input.LogicalName)
	}
}
```
