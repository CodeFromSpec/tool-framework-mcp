[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@U2SmPtOwjQYJCsOrURkBXcVUS3g)

## Package

```go
package mcpvalidatespecs
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
```

## Structs

```go
package mcpvalidatespecs

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"

// StalenessEntry describes a single artifact whose hash is missing, stale, or malformed.
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

// ValidationReport is the result produced by MCPValidateSpecs.
type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}
```

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree rooted at code-from-spec/,
// validates format rules, detects dependency cycles, and checks whether
// each artifact file is present and up-to-date.
//
// It always returns a report. Problems are collected inside the report;
// no error is returned.
//
// StalenessEntry.Status is one of:
//   - "missing"       — the artifact file does not exist.
//   - "stale"         — the artifact file exists but its embedded hash does not match.
//   - "malformed tag" — the artifact file exists but has no artifact tag or the tag cannot be parsed.
//
// Entries whose hash matches are not included in ValidationReport.Staleness.
// ValidationReport.Cycles is the flat list of logical names involved in
// non-convergence during ranking, as returned by NodeRankCompute.
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
		fmt.Println("All nodes are valid and up-to-date.")
		return
	}

	for _, fe := range report.FormatErrors {
		fmt.Printf("Format error — node: %s | rule: %s | detail: %s\n", fe.Node, fe.Rule, fe.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("Cycle detected: %s\n", name)
	}

	for _, s := range report.Staleness {
		fmt.Printf("Staleness [%s] — node: %s | artifact: %s | rank: %d | detail: %s\n",
			s.Status, s.Node, s.ArtifactPath, s.Rank, s.Detail)
	}
}
```
