[//]: # (code-from-spec: ROOT/golang/interfaces/chain/resolver@VD-nH9QO_ui7Tc9A-B1j38D8TMU)

# Package `chainresolver`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
```

Package `chainresolver` resolves the ordered chain of spec nodes required for artifact generation or chain hash computation, given a target logical name.

---

## Structs

```go
package chainresolver

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ChainItem represents a single node entry within a resolved chain.
type ChainItem struct {
	// LogicalName is the logical name of the node.
	LogicalName string

	// FilePath is the CFS path to the node's file.
	FilePath pathutils.PathCfs

	// Qualifier is an optional qualifier for the entry.
	// Empty string when absent.
	Qualifier string
}

// Chain holds the fully resolved chain for a target node.
// Items are ordered as described in the chain assembly specification.
type Chain struct {
	// Ancestors is the ordered list of ancestor nodes, from root down
	// to (but not including) the target node.
	Ancestors []*ChainItem

	// Dependencies is the list of entries from the target's depends_on,
	// sorted alphabetically by file path then by qualifier, each with
	// its resolved file path and an optional qualifier.
	Dependencies []*ChainItem

	// External is the list of external file references from the target's
	// frontmatter, sorted alphabetically by path, including fragment
	// declarations when present.
	External []*frontmatter.FrontmatterExternal

	// Target is the target node itself.
	Target *ChainItem

	// Input is the target's input artifact. Nil when absent.
	Input *ChainItem
}
```

---

## Error Sentinels

```go
package chainresolver

import "errors"

// ErrUnreadableFrontmatter is returned when a node's frontmatter
// cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's
// output id does not match any declared output.
var ErrUnresolvableArtifact = errors.New("unresolvable artifact reference")
```

---

## Functions

```go
package chainresolver

// ChainResolve returns the resolved chain for the given target logical
// name. The chain contains ancestors, dependencies, external references,
// the target node, and optionally an input artifact — in the order
// required for context assembly or chain hash computation.
//
// Chain assembly order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — entries from the target's depends_on, sorted
//     alphabetically by file path then by qualifier, each with its
//     resolved file path and an optional qualifier.
//  3. External — files from the target's external, sorted alphabetically
//     by path, including fragment declarations when present.
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
	// Resolve the chain for a target logical name.
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		if errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
			log.Fatal("could not parse a node's frontmatter")
		}
		if errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
			log.Fatal("an ARTIFACT/ reference could not be resolved")
		}
		if errors.Is(err, frontmatter.ErrFileUnreadable) {
			log.Fatal("a spec file could not be read")
		}
		if errors.Is(err, frontmatter.ErrMalformedYAML) {
			log.Fatal("a spec file contains invalid YAML")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	// Print the chain assembly order.
	fmt.Println("=== Ancestors ===")
	for _, item := range chain.Ancestors {
		fmt.Printf("  %s -> %s\n", item.LogicalName, item.FilePath.Value)
	}

	fmt.Println("=== Dependencies ===")
	for _, item := range chain.Dependencies {
		qualifier := item.Qualifier
		if qualifier == "" {
			qualifier = "(none)"
		}
		fmt.Printf("  %s -> %s [qualifier: %s]\n", item.LogicalName, item.FilePath.Value, qualifier)
	}

	fmt.Println("=== External ===")
	for _, ext := range chain.External {
		fmt.Printf("  %s (%d fragments)\n", ext.Path, len(ext.Fragments))
	}

	fmt.Println("=== Target ===")
	fmt.Printf("  %s -> %s\n", chain.Target.LogicalName, chain.Target.FilePath.Value)

	fmt.Println("=== Input ===")
	if chain.Input != nil {
		fmt.Printf("  %s -> %s\n", chain.Input.LogicalName, chain.Input.FilePath.Value)
	} else {
		fmt.Println("  (none)")
	}
}
```
