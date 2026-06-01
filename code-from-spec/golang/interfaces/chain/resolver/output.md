[//]: # (code-from-spec: ROOT/golang/interfaces/chain/resolver@NRWKzAOgBl0Y_1sQMgT-e5rU1DE)

# Package `chainresolver`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
```

Package `chainresolver` resolves the ordered chain of spec nodes required to assemble context for a target logical name.

---

## Structs

```go
package chainresolver

// ChainItem represents a single node entry in a resolved chain.
type ChainItem struct {
	LogicalName string
	FilePath    pathutils.PathCfs
	Qualifier   *string
}

// Chain holds the fully resolved chain for a target logical name.
type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	External     []*frontmatter.FrontmatterExternal
	Target       *ChainItem
	Input        *ChainItem
}
```

---

## Error Sentinels

```go
package chainresolver

import "errors"

// ErrUnreadableFrontmatter is returned when a node's frontmatter cannot
// be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's
// output id does not match any declared output.
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")
```

---

## Functions

```go
package chainresolver

// ChainResolve returns the chain for the given target logical name.
//
// The chain is assembled in the following order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — entries from the target's depends_on, sorted
//     alphabetically by file path then by qualifier.
//  3. External — files from the target's external, sorted alphabetically
//     by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
//
// Errors:
//   - ErrUnreadableFrontmatter: a node's frontmatter cannot be parsed.
//   - ErrUnresolvableArtifact: an ARTIFACT/ reference's output id does
//     not match any declared output.
//   - (LogicalNames.*): propagated from LogicalNameToPath,
//     LogicalNameGetParent.
//   - (Frontmatter.*): propagated from FrontmatterParse.
func ChainResolve(targetLogicalName string) (*Chain, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
)

func main() {
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		if errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
			log.Fatal("could not parse a node's frontmatter")
		}
		if errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
			log.Fatal("an ARTIFACT/ reference could not be resolved")
		}
		log.Fatalf("chain resolution failed: %v", err)
	}

	fmt.Println("target:", chain.Target.LogicalName)

	for _, ancestor := range chain.Ancestors {
		fmt.Println("ancestor:", ancestor.LogicalName)
	}

	for _, dep := range chain.Dependencies {
		qualifier := ""
		if dep.Qualifier != nil {
			qualifier = *dep.Qualifier
		}
		fmt.Printf("dependency: %s qualifier=%s\n", dep.LogicalName, qualifier)
	}

	for _, ext := range chain.External {
		fmt.Println("external:", ext.Path)
	}

	if chain.Input != nil {
		fmt.Println("input:", chain.Input.LogicalName)
	}

	_ = frontmatter.FrontmatterExternal{}
}
```
