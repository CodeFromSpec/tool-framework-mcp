[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@C13zOZdV9QG-aYyAd0BHyG0FtXs)

# Interface: `spectreevalidate`

**Package:** `package spectreevalidate`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"`

---

## Structs

```go
// SpecTreeValidateInput holds a single discovered node's logical name along
// with its parsed frontmatter and parsed node body, as produced by the
// frontmatter and parsenode packages respectively.
type SpecTreeValidateInput struct {
    LogicalName string
    Frontmatter *frontmatter.Frontmatter
    Node        *parsenode.Node
}

// FormatError describes a single format violation found in a node.
// Node is the logical name of the offending node. Rule identifies
// which validation rule was violated. Detail provides a human-readable
// explanation of the specific violation.
type FormatError struct {
    Node   string
    Rule   string
    Detail string
}
```

---

## Functions

```go
// SpecTreeValidate validates the full set of discovered nodes against the
// spec tree format rules. It accepts all entries at once so that cross-node
// rules (such as parent/child/leaf relationships) can be evaluated.
//
// A node is considered to have children if any other entry in the input list
// has a logical name that starts with the node's logical name followed by "/".
// For example, given entries "ROOT/a" and "ROOT/a/b", "ROOT/a" has children.
// A node is a leaf if no entry has a logical name that starts with the node's
// logical name followed by "/".
//
// Returns a list of FormatError values describing every violation found across
// all entries. Returns an empty list if all nodes are valid.
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
    logicalNames := []string{
        "ROOT/a",
        "ROOT/a/b",
    }

    var entries []*spectreevalidate.SpecTreeValidateInput

    for _, name := range logicalNames {
        cfsPath := &pathutils.PathCfs{Value: "code-from-spec/.../_node.md"} // resolved from logical name

        fm, err := frontmatter.FrontmatterParse(cfsPath)
        if err != nil {
            log.Fatalf("failed to parse frontmatter for %s: %v", name, err)
        }

        node, err := parsenode.NodeParse(name)
        if err != nil {
            log.Fatalf("failed to parse node %s: %v", name, err)
        }

        entries = append(entries, &spectreevalidate.SpecTreeValidateInput{
            LogicalName: name,
            Frontmatter: fm,
            Node:        node,
        })
    }

    errs := spectreevalidate.SpecTreeValidate(entries)
    if len(errs) == 0 {
        fmt.Println("all nodes are valid")
        return
    }

    for _, e := range errs {
        fmt.Printf("node: %s | rule: %s | detail: %s\n", e.Node, e.Rule, e.Detail)
    }
}
```
