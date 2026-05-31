[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@c8-GesIoWsU5auTsXAJQq7wZ7Uk)

# Package `mcpvalidatespecs`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
```

Scans the entire spec tree starting from `code-from-spec/` and returns a
`ValidationReport` describing format errors, dependency cycles, and output
staleness. Never raises an error — problems are collected into the report.

---

## Structs

```go
package mcpvalidatespecs

// StalenessEntry describes a single output file whose artifact tag is missing,
// unparseable, or whose hash does not match the current chain hash.
type StalenessEntry struct {
	// Node is the logical name of the node that owns the output.
	Node string

	// OutputID is the id field from the node's outputs frontmatter.
	OutputID string

	// ArtifactPath is the relative path to the generated file.
	ArtifactPath string

	// Status is one of "missing", "stale", or "malformed tag".
	Status string

	// Detail provides additional context about the staleness condition.
	Detail string

	// Rank is the dependency rank of the node as returned by NodeRankCompute.
	// Entries with equal rank have no dependency between them and can be
	// processed in parallel.
	Rank int
}

// ValidationReport is the result of MCPValidateSpecs. It collects all
// problems found across the spec tree.
type ValidationReport struct {
	// FormatErrors lists nodes that violate spec-tree structural rules.
	FormatErrors []*spectreevalidate.FormatError

	// Cycles is a flat list of logical names involved in non-convergence
	// during ranking, as returned by NodeRankCompute.
	Cycles []string

	// Staleness lists output files whose artifact tag is missing, stale,
	// or malformed. Files whose hash matches are not included.
	Staleness []*StalenessEntry
}
```

---

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree rooted at code-from-spec/,
// validates node structure, computes dependency ranks, and checks every
// declared output file for a current artifact tag.
//
// It always returns a report. Callers should inspect FormatErrors, Cycles,
// and Staleness to determine whether any problems were found.
func MCPValidateSpecs() *ValidationReport
```

---

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
)

func main() {
	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
		fmt.Println("Spec tree is valid and all outputs are up to date.")
		return
	}

	for _, fe := range report.FormatErrors {
		fmt.Printf("format error — node: %s  rule: %s  detail: %s\n", fe.Node, fe.Rule, fe.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("cycle — node: %s\n", name)
	}

	for _, se := range report.Staleness {
		fmt.Printf("staleness — node: %s  output: %s  path: %s  status: %s  detail: %s  rank: %d\n",
			se.Node, se.OutputID, se.ArtifactPath, se.Status, se.Detail, se.Rank)
	}
}
```
