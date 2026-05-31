[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@iDG9mI29zbbvCuDo_HWjry2wIP8)

# Package `mcpvalidatespecs`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
```

Package `mcpvalidatespecs` implements the `validate_specs` MCP tool. It scans the entire spec tree starting from `code-from-spec/`, validates all nodes, checks artifact staleness, and returns a structured report. The tool never raises an error — all problems are collected in the report.

---

## Structs

```go
package mcpvalidatespecs

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

// StalenessEntry describes a single output artifact that is missing,
// stale, or has a malformed artifact tag.
//
// Artifacts whose hash matches are not included in the report.
type StalenessEntry struct {
	// Node is the logical name of the spec tree node that owns the output.
	Node string

	// OutputID is the id of the output as declared in the node's frontmatter.
	OutputID string

	// ArtifactPath is the relative path to the artifact file.
	ArtifactPath string

	// Status describes the staleness condition. One of:
	//   "missing"       — the file does not exist.
	//   "stale"         — the file exists but the hash does not match.
	//   "malformed tag" — the file exists but has no artifact tag or the
	//                     tag cannot be parsed.
	Status string

	// Detail provides a human-readable explanation of the staleness condition.
	Detail string

	// Rank is the node's rank as returned by NodeRankCompute. Entries with
	// equal rank have no dependency between them and can be processed in
	// parallel.
	Rank int
}

// ValidationReport is the result returned by MCPValidateSpecs.
// It aggregates all discovered problems across the spec tree.
type ValidationReport struct {
	// FormatErrors is the list of format rule violations found during
	// spec tree validation.
	FormatErrors []*spectreevalidate.FormatError

	// Cycles is a flat list of logical names involved in non-convergence
	// during ranking, as returned by NodeRankCompute.
	Cycles []string

	// Staleness is the list of output artifacts that are missing, stale,
	// or have a malformed artifact tag.
	Staleness []*StalenessEntry
}
```

---

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the entire spec tree starting from
// "code-from-spec/", validates all nodes, and checks each declared
// output artifact for staleness.
//
// It never returns an error. All discovered problems — format errors,
// ranking cycles, and stale or missing artifacts — are collected and
// returned in a ValidationReport.
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
		fmt.Println("spec tree is valid and all artifacts are up to date")
		return
	}

	for _, fe := range report.FormatErrors {
		fmt.Printf("format error: node=%s rule=%s detail=%s\n", fe.Node, fe.Rule, fe.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("cycle: node=%s\n", name)
	}

	for _, se := range report.Staleness {
		fmt.Printf("staleness: node=%s output=%s path=%s status=%s rank=%d detail=%s\n",
			se.Node, se.OutputID, se.ArtifactPath, se.Status, se.Rank, se.Detail)
	}
}
```
