[//]: # (code-from-spec: ROOT/golang/interfaces/chain/resolver@cmDJkI3yJkxYXdz3v5EtYfP7MlY)

# Package `chainresolver`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
```

Resolves the ordered chain of spec nodes required to assemble context for artifact generation or to compute the chain hash.

---

## Structs

```go
package chainresolver

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ChainItem represents a single node position within a resolved chain.
type ChainItem struct {
	// LogicalName is the logical name of the node.
	LogicalName string

	// FilePath is the CFS-format path to the node's spec file.
	FilePath *pathutils.PathCfs

	// Qualifier is an optional qualifier string, used for dependency entries.
	Qualifier *string
}

// Chain is the fully resolved chain for a target node, in assembly order.
type Chain struct {
	// Ancestors holds the nodes from root down to (but not including) the target.
	Ancestors []*ChainItem

	// Dependencies holds entries from the target's depends_on, sorted
	// alphabetically by file path then by qualifier.
	Dependencies []*ChainItem

	// External holds external file references from the target's frontmatter,
	// sorted alphabetically by path, including any fragment declarations.
	External []*frontmatter.FrontmatterExternal

	// Target is the target node itself.
	Target *ChainItem

	// Input is the target's input artifact, if present.
	Input *ChainItem
}
```

---

## Error Sentinels

```go
package chainresolver

import "errors"

// ErrUnreadableFrontmatter is returned when a node's frontmatter cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's output id
// does not match any declared output in the referenced node.
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
//  3. External — files from the target's external field, sorted
//     alphabetically by path, including fragment declarations when present.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
//
// Errors:
//   - ErrUnreadableFrontmatter: a node's frontmatter cannot be parsed.
//   - ErrUnresolvableArtifact: an ARTIFACT/ reference's output id does not
//     match any declared output.
//   - (LogicalNames.*): propagated from LogicalNameToPath, LogicalNameGetParent.
//   - (Frontmatter.*): propagated from FrontmatterParse.
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
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
)

func main() {
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		log.Fatalf("ChainResolve: %v", err)
	}

	fmt.Println("=== Ancestors ===")
	for _, item := range chain.Ancestors {
		fmt.Printf("  %s -> %s\n", item.LogicalName, item.FilePath.Value)
	}

	fmt.Println("=== Dependencies ===")
	for _, item := range chain.Dependencies {
		qualifier := "<none>"
		if item.Qualifier != nil {
			qualifier = *item.Qualifier
		}
		fmt.Printf("  %s -> %s (qualifier: %s)\n", item.LogicalName, item.FilePath.Value, qualifier)
	}

	fmt.Println("=== External ===")
	for _, ext := range chain.External {
		fmt.Printf("  %s (%d fragments)\n", ext.Path, len(ext.Fragments))
		for _, frag := range ext.Fragments {
			fmt.Printf("    fragment — description: %q, lines: %s, hash: %s\n",
				frag.Description, frag.Lines, frag.Hash)
		}
	}

	fmt.Println("=== Target ===")
	fmt.Printf("  %s -> %s\n", chain.Target.LogicalName, chain.Target.FilePath.Value)

	fmt.Println("=== Input ===")
	if chain.Input != nil {
		fmt.Printf("  %s -> %s\n", chain.Input.LogicalName, chain.Input.FilePath.Value)
	} else {
		fmt.Println("  <none>")
	}

	// Sentinel errors can be checked with errors.Is:
	//
	//   _, err := chainresolver.ChainResolve("ROOT/nonexistent")
	//   if errors.Is(err, chainresolver.ErrUnreadableFrontmatter) { ... }
	//   if errors.Is(err, chainresolver.ErrUnresolvableArtifact) { ... }
	//   if errors.Is(err, frontmatter.ErrMalformedYAML) { ... }
	_ = frontmatter.ErrMalformedYAML // imported for documentation purposes
}
```
