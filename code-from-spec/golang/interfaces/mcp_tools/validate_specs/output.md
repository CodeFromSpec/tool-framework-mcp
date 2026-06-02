[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@4MZmmMxRNhroyo2ubyFAamZmQ1M)

# Package `mcpvalidatespecs`

```go
package mcpvalidatespecs
```

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"`

## Struct Definitions

```go
package mcpvalidatespecs

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"

// StalenessEntry describes a single artifact whose staleness check failed.
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

// ValidationReport holds the full result of a spec tree validation run.
type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}
```

## Function Signatures

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree starting from code-from-spec/,
// validates format, computes ranks, and checks artifact staleness. Always
// returns a report — never returns an error. Problems are collected in the
// report fields.
//
// StalenessEntry.Status is one of:
//   - "missing"        — the artifact file does not exist
//   - "stale"          — the artifact file exists but its hash does not match
//   - "malformed tag"  — the artifact file exists but has no parseable artifact tag
//
// Entries where the hash matches are not included in the Staleness list.
// Cycles is a flat list of logical names involved in non-convergence during ranking.
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
		fmt.Println("spec tree is valid and all artifacts are fresh")
		return
	}

	for _, e := range report.FormatErrors {
		fmt.Printf("format error: node=%s rule=%s detail=%s\n", e.Node, e.Rule, e.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("cycle: %s\n", name)
	}

	for _, s := range report.Staleness {
		fmt.Printf("staleness: node=%s path=%s status=%s rank=%d detail=%s\n",
			s.Node, s.ArtifactPath, s.Status, s.Rank, s.Detail)
	}
}
```
