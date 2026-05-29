[//]: # (code-from-spec: ROOT/golang/interfaces/chain/resolver@i9BWV_qvwUbEMHETkB2qfak1Qeo)

# Interface: `chainresolver`

**Package:** `package chainresolver`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"`

---

## Structs

```go
// ChainItem represents a single node entry within a resolved chain.
// It holds the logical name, the CFS file path to the node's spec
// file, and an optional qualifier (used when the item originates
// from an ARTIFACT/ reference that targets a specific output id).
type ChainItem struct {
    LogicalName string
    FilePath    *pathutils.PathCfs
    Qualifier   *string
}

// Chain holds the fully resolved chain for a target logical name.
// It contains the ordered lists of ancestors, dependencies, external
// files, the target itself, and an optional input artifact.
//
// Assembly order mirrors the order in which a downstream tool should
// concatenate context:
//  1. Ancestors — root down to (but not including) the target.
//  2. Dependencies — target's depends_on, sorted alphabetically by
//     file path then qualifier.
//  3. External — target's external files, sorted alphabetically by
//     path.
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
    // ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's
    // output id does not match any output declared in the referenced
    // node's frontmatter.
    ErrUnresolvableArtifact = errors.New("unresolvable artifact")
)
```

---

## Functions

```go
// ChainResolve builds and returns the full chain for the given target
// logical name. It walks the ancestor hierarchy, collects depends_on
// entries, gathers external file references, and resolves the optional
// input artifact.
//
// The returned Chain fields are ordered as described in the Chain struct
// documentation.
//
// Returns an error if:
//   - the target logical name is invalid or cannot be converted to a
//     path (errors propagated from LogicalNameToPath / LogicalNameGetParent).
//   - any node's frontmatter cannot be parsed (frontmatter parse errors
//     are propagated).
//   - an ARTIFACT/ reference in depends_on specifies an output id that
//     does not match any declared output in the referenced node's
//     frontmatter (ErrUnresolvableArtifact).
func ChainResolve(target_logical_name string) (*Chain, error)
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
    chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/resolver")
    if err != nil {
        log.Fatalf("failed to resolve chain: %v", err)
    }

    fmt.Println("Target:", chain.Target.LogicalName)
    fmt.Println("Target path:", chain.Target.FilePath.Value)

    fmt.Println("Ancestors:")
    for _, a := range chain.Ancestors {
        fmt.Printf("  %s (%s)\n", a.LogicalName, a.FilePath.Value)
    }

    fmt.Println("Dependencies:")
    for _, d := range chain.Dependencies {
        qualifier := "<none>"
        if d.Qualifier != nil {
            qualifier = *d.Qualifier
        }
        fmt.Printf("  %s (%s) qualifier=%s\n", d.LogicalName, d.FilePath.Value, qualifier)
    }

    fmt.Println("External files:")
    for _, e := range chain.External {
        fmt.Printf("  %s (%d fragments)\n", e.Path, len(e.Fragments))
    }

    if chain.Input != nil {
        fmt.Println("Input:", chain.Input.LogicalName, chain.Input.FilePath.Value)
    }
}
```
