[//]: # (code-from-spec: ROOT/golang/interfaces/chain/chain_resolver@HeolNsy_r8eKIXV86pVvz073dS8)

# Interface: `chainresolver`

## Package

```go
package chainresolver
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainresolver"
```

---

## Struct Definitions

```go
// ChainItem represents a single node in the resolved chain, identified
// by its logical name and the CFS path to its spec file. An optional
// qualifier is set when the item was referenced via an ARTIFACT/ prefix
// and points to a specific output id.
type ChainItem struct {
	LogicalName string
	FilePath    *pathutils.PathCfs
	Qualifier   *string
}

// Chain holds the fully resolved chain for a target node. It contains
// the ordered context needed by downstream tools to assemble the prompt
// or compute the chain hash.
//
// Assembly order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — from the target's depends_on, sorted alphabetically
//     by file path then by qualifier.
//  3. External — from the target's external, sorted alphabetically by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
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
var (
	// ErrInvalidLogicalName is returned when the target logical name
	// cannot be converted to a file path or its parent cannot be derived.
	// This sentinel wraps the underlying logical-name error.
	ErrInvalidLogicalName = errors.New("invalid logical name")

	// ErrUnreadableFrontmatter is returned when a node's spec file cannot
	// be opened or its frontmatter cannot be parsed.
	ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

	// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference
	// specifies an output id that does not appear in the node's declared
	// outputs list.
	ErrUnresolvableArtifact = errors.New("unresolvable artifact")
)
```

---

## Functions

```go
// ChainResolve builds and returns the full Chain for the given target
// logical name.
//
// The returned Chain contains ancestors (root down to but not including
// the target), dependencies (from the target's depends_on, sorted
// alphabetically by file path then qualifier), external entries (from
// the target's external, sorted alphabetically by path), the target
// itself, and optionally the input artifact.
//
// Possible errors:
//   - ErrInvalidLogicalName — propagated from LogicalNameToPath or
//     LogicalNameGetParent when the logical name is malformed.
//   - ErrUnreadableFrontmatter — a node's frontmatter cannot be opened
//     or parsed (wraps the underlying frontmatter error).
//   - ErrUnresolvableArtifact — an ARTIFACT/ reference's output id does
//     not match any declared output in the referenced node's frontmatter.
func ChainResolve(target_logical_name string) (*Chain, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/chain_resolver")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Target:", chain.Target.LogicalName)
	fmt.Println("File:  ", chain.Target.FilePath.Value)

	fmt.Printf("Ancestors (%d):\n", len(chain.Ancestors))
	for _, a := range chain.Ancestors {
		fmt.Println(" -", a.LogicalName)
	}

	fmt.Printf("Dependencies (%d):\n", len(chain.Dependencies))
	for _, d := range chain.Dependencies {
		q := "<none>"
		if d.Qualifier != nil {
			q = *d.Qualifier
		}
		fmt.Printf("  - %s (qualifier: %s)\n", d.LogicalName, q)
	}

	fmt.Printf("External (%d):\n", len(chain.External))
	for _, e := range chain.External {
		fmt.Printf("  - %s (%d fragments)\n", e.Path, len(e.Fragments))
	}

	if chain.Input != nil {
		fmt.Println("Input:", chain.Input.LogicalName)
	}
}
```
