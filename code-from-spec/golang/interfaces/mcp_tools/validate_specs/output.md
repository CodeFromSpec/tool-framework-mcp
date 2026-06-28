[//]: # (code-from-spec: SPEC/golang/interfaces/mcp_tools/validate_specs@oqpjBG_pX5rAIBu_plNvU9teKBI)

# Package `mcpvalidatespecs`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpvalidatespecs`

## Types

```go
package mcpvalidatespecs

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectreevalidate"
)

// StalenessEntry describes the staleness status of a single output artifact.
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

// ValidationReport holds the full result of a spec-tree validation run.
type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}
```

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree starting from code-from-spec/,
// validates all nodes against format rules, detects dependency cycles, and
// checks whether each output artifact is up to date. Always returns a report —
// never returns an error. Problems are collected in the report fields.
//
// StalenessEntry.Status is one of:
//   - "missing"       — file does not exist.
//   - "stale"         — hash mismatch.
//   - "malformed tag" — file exists but has no artifact tag or the tag cannot be parsed.
//
// Entries whose hash matches are not included in Staleness.
// Cycles contains logical names involved in non-convergence during ranking.
// StalenessEntry.Rank is the rank from NodeRankCompute; entries with equal rank
// have no dependency between them and can be processed in parallel.
func MCPValidateSpecs() *ValidationReport
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpvalidatespecs"
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
		fmt.Printf("Cycle detected involving: %s\n", name)
	}

	for _, s := range report.Staleness {
		fmt.Printf("Staleness — Node: %s | Path: %s | Status: %s | Rank: %d | Detail: %s\n",
			s.Node, s.ArtifactPath, s.Status, s.Rank, s.Detail)
	}
}
```
