[//]: # (code-from-spec: SPEC/golang/interfaces/chain/resolver@cWUbI5sDEWlpFKoZKgZuYnJk3g0)

# Package `chainresolver`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver`

## Types

```go
package chainresolver

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

// ChainItem represents a single node in the resolved chain.
type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               pathutils.PathCfs
	Qualifier              string
}

// Chain is the fully resolved chain for a target logical name.
type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
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

// ChainResolve returns the chain for the given target logical name.
// The chain contains ancestors (root down to but not including the target),
// dependencies (sorted alphabetically by logical name), the target itself,
// and optionally the target's input.
func ChainResolve(targetLogicalName string) (*Chain, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Target:", chain.Target.UnqualifiedLogicalName)
	fmt.Println("Target file:", chain.Target.FilePath.Value)

	fmt.Println("Ancestors:")
	for _, a := range chain.Ancestors {
		fmt.Println(" ", a.UnqualifiedLogicalName, a.FilePath.Value)
	}

	fmt.Println("Dependencies:")
	for _, d := range chain.Dependencies {
		qualifier := ""
		if d.Qualifier != "" {
			qualifier = "(" + d.Qualifier + ")"
		}
		fmt.Println(" ", d.UnqualifiedLogicalName+qualifier, d.FilePath.Value)
	}

	if chain.Input != nil {
		fmt.Println("Input:", chain.Input.UnqualifiedLogicalName, chain.Input.FilePath.Value)
	}
}
```
