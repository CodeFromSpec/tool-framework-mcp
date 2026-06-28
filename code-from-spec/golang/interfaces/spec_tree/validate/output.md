[//]: # (code-from-spec: SPEC/golang/interfaces/spec_tree/validate@VlNXBu7HDK8tfQBYKtK6ahWVm7w)

# Package `spectreevalidate`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectreevalidate`

## Types

```go
package spectreevalidate

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
)

// SpecTreeValidateInput holds a discovered node together with its parsed
// frontmatter and body structure, as inputs to validation.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError describes a single validation violation found in a node.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}
```

## Functions

```go
package spectreevalidate

// SpecTreeValidate validates all discovered nodes against the spec-tree
// format rules. entries is the full set of nodes with their parsed
// frontmatter and body. allDirs is the list of all subdirectory paths
// found under code-from-spec/. Returns a list of FormatErrors (empty
// when all nodes are valid).
func SpecTreeValidate(entries []*SpecTreeValidateInput, allDirs []string) []*FormatError
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectreevalidate"
)

func main() {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "SPEC/payments/fees",
			Frontmatter: &frontmatter.Frontmatter{
				Output: "src/fees.go",
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{
					Heading: "payments/fees",
				},
				Public: &parsenode.NodeSection{
					Heading: "Public",
				},
			},
		},
		{
			LogicalName: "SPEC/payments/fees/calculator",
			Frontmatter: &frontmatter.Frontmatter{
				Output: "src/calculator.go",
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{
					Heading: "payments/fees/calculator",
				},
				Public: &parsenode.NodeSection{
					Heading: "Public",
				},
			},
		},
	}

	allDirs := []string{
		"code-from-spec/payments",
		"code-from-spec/payments/fees",
		"code-from-spec/payments/fees/calculator",
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
