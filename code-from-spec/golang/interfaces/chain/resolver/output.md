# code-from-spec: ROOT/golang/interfaces/chain/resolver@HX0dS1xcAcsBeMpISQVe4hpRz3c

# Package `chainresolver`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver`

---

## Structs

```go
package chainresolver

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ChainItem represents a single node position in a resolved chain.
type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               pathutils.PathCfs
	Qualifier              *string
}

// Chain holds the fully resolved chain for a target node.
type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	Target       *ChainItem
	Input        *ChainItem
}
```

---

## Error Sentinels

```go
package chainresolver

import "errors"

var ErrUnreadableFrontmatter = errors.New("a node's frontmatter cannot be parsed")
var ErrUnresolvableArtifact  = errors.New("an ARTIFACT/ reference cannot be resolved")
```

---

## Functions

```go
package chainresolver

// ChainResolve returns the chain for the given target logical name.
// The chain contains ancestors (root down to but not including the target),
// dependencies (all depends_on entries sorted alphabetically), the target
// itself, and optionally the target's input node.
func ChainResolve(targetLogicalName string) (*Chain, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("SPEC/payments/invoices")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Target:", chain.Target.UnqualifiedLogicalName)
	fmt.Println("Target file:", chain.Target.FilePath.Value)

	for _, ancestor := range chain.Ancestors {
		fmt.Println("Ancestor:", ancestor.UnqualifiedLogicalName)
	}

	for _, dep := range chain.Dependencies {
		qualifier := ""
		if dep.Qualifier != nil {
			qualifier = " (" + *dep.Qualifier + ")"
		}
		fmt.Printf("Dependency: %s%s -> %s\n", dep.UnqualifiedLogicalName, qualifier, dep.FilePath.Value)
	}

	if chain.Input != nil {
		fmt.Println("Input:", chain.Input.UnqualifiedLogicalName)
	}
}
```
