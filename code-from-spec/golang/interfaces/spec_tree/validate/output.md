[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@yRa3zlQHmAERCUMElnAaYPh72Ws)

# Package `spectreevalidate`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate`

---

## Structs

```go
package spectreevalidate

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
)

// SpecTreeValidateInput holds a discovered node with its parsed frontmatter and body.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError describes a single validation violation for a node.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}
```

---

## Functions

```go
package spectreevalidate

// SpecTreeValidate validates the full set of discovered nodes against format rules.
// entries is the complete list of nodes with their parsed frontmatter and body.
// allDirs is the list of all subdirectory paths found under code-from-spec/.
// Returns a list of FormatErrors; the list is empty when all nodes are valid.
func SpecTreeValidate(entries []*SpecTreeValidateInput, allDirs []string) []*FormatError
```

---

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

func main() {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "SPEC/payments/fees",
			Frontmatter: &frontmatter.Frontmatter{
				Output: "ARTIFACT/payments/fees/output.md",
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{Heading: "payments/fees"},
			},
		},
		{
			LogicalName: "SPEC/payments/fees/calculate",
			Frontmatter: &frontmatter.Frontmatter{
				Output: "ARTIFACT/payments/fees/calculate/output.md",
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{Heading: "payments/fees/calculate"},
			},
		},
	}

	allDirs := []string{
		"code-from-spec/payments",
		"code-from-spec/payments/fees",
		"code-from-spec/payments/fees/calculate",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) == 0 {
		fmt.Println("All nodes are valid.")
		return
	}

	for _, e := range errs {
		fmt.Printf("Node: %s | Rule: %s | Detail: %s\n", e.Node, e.Rule, e.Detail)
	}
}
```
