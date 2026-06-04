[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@6HCJBYFHKMXMBZHMEyyNAoxiFjU)

# Package `mcpvalidatespecs`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
```

## Structs

```go
package mcpvalidatespecs

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"

type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}
```

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree starting from code-from-spec/
// and returns a ValidationReport. It never returns an error — all problems
// are collected in the report.
//
// FormatErrors contains structural validation failures for spec nodes.
// Cycles contains logical names involved in non-convergence during ranking.
// Staleness contains entries for artifacts whose hash is missing, mismatched,
// or whose tag cannot be parsed. Entries with a matching hash are omitted.
//
// StalenessEntry.Status is one of:
//   - "missing"       — file does not exist.
//   - "stale"         — hash mismatch.
//   - "malformed tag" — file exists but has no artifact tag or the tag cannot be parsed.
//
// StalenessEntry.Rank is the rank from NodeRankCompute. Entries with equal
// rank have no dependency between them and can be processed in parallel.
func MCPValidateSpecs() *ValidationReport
```

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
		fmt.Println("Spec tree is valid and all artifacts are up to date.")
		return
	}

	for _, e := range report.FormatErrors {
		fmt.Printf("Format error — Node: %s | Rule: %s | Detail: %s\n", e.Node, e.Rule, e.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("Cycle detected — Node: %s\n", name)
	}

	for _, s := range report.Staleness {
		fmt.Printf("Staleness [%s] (rank %d) — %s: %s\n", s.Status, s.Rank, s.Node, s.Detail)
	}
}
```
