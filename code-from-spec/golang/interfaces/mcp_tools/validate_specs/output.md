[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@p869O8hj7vSQxb0ZPT9fWVj68k0)

# Package `mcpvalidatespecs`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
```

Scans the entire spec tree starting from `code-from-spec/` and returns a
report describing format errors, dependency cycles, and staleness of generated
artifacts. Never returns an error — all problems are collected in the report.

---

## Structs

```go
package mcpvalidatespecs

// StalenessEntry describes a single output file whose artifact tag is missing,
// malformed, or does not match the current chain hash.
type StalenessEntry struct {
	// Node is the logical name of the owning node (e.g. "ROOT/foo/bar").
	Node string

	// OutputID is the output id declared in the node's frontmatter (e.g. "interface").
	OutputID string

	// ArtifactPath is the relative path of the output file from the project root.
	ArtifactPath string

	// Status is one of "missing", "stale", or "malformed tag".
	Status string

	// Detail provides additional context about the staleness condition.
	Detail string

	// Rank is the topological rank of the node as returned by NodeRankCompute.
	// Entries with equal rank have no dependency between them and can be
	// processed in parallel.
	Rank int
}

// ValidationReport is the result returned by MCPValidateSpecs.
type ValidationReport struct {
	// FormatErrors lists all node format violations found during spec tree
	// validation.
	FormatErrors []*spectreevalidate.FormatError

	// Cycles is a flat list of logical names involved in non-convergence during
	// ranking, as returned by NodeRankCompute.
	Cycles []string

	// Staleness lists all output files that are missing, stale, or have a
	// malformed artifact tag. Entries where the hash matches are not included.
	Staleness []*StalenessEntry
}
```

---

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree starting from code-from-spec/,
// validates node format, computes dependency ranks, and checks every declared
// output file for a current and matching artifact tag.
//
// It always returns a report. Problems are collected inside the report rather
// than returned as errors.
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
		fmt.Println("All nodes are valid and all artifacts are up to date.")
		return
	}

	for _, fe := range report.FormatErrors {
		fmt.Printf("format error  node: %s  rule: %s  detail: %s\n", fe.Node, fe.Rule, fe.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("cycle node: %s\n", name)
	}

	for _, se := range report.Staleness {
		fmt.Printf(
			"staleness  node: %s  output: %s  path: %s  status: %s  rank: %d  detail: %s\n",
			se.Node, se.OutputID, se.ArtifactPath, se.Status, se.Rank, se.Detail,
		)
	}
}
```
