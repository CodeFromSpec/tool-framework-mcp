[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@VFqk-LjCdTrm3pmg4bTt_DQgaMc)

# Package `mcpvalidatespecs`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
```

Package `mcpvalidatespecs` implements the `validate_specs` MCP tool. It scans the entire spec tree, validates format rules, detects ranking cycles, and checks artifact staleness. It always returns a report — problems are collected rather than surfaced as errors.

---

## Structs

```go
package mcpvalidatespecs

// StalenessEntry describes a single output artifact whose staleness
// check did not pass. Entries where the hash matches are not included.
type StalenessEntry struct {
	Node         string
	OutputID     string
	ArtifactPath string
	Status       string
	Rank         int
	Detail       string
}

// ValidationReport is the result of MCPValidateSpecs. All three fields
// are populated (possibly empty) regardless of any problems found.
type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}
```

`StalenessEntry.Status` is one of:

- `"missing"` — the artifact file does not exist.
- `"stale"` — the file exists but its embedded hash does not match the current chain hash.
- `"malformed tag"` — the file exists but contains no artifact tag or the tag cannot be parsed.

`StalenessEntry.Rank` is the rank produced by `NodeRankCompute`. Entries with equal rank have no dependency between them and can be regenerated in parallel.

`ValidationReport.Cycles` is a flat list of logical names involved in non-convergence during ranking, as returned by `NodeRankCompute`.

---

## Functions

```go
package mcpvalidatespecs

// MCPValidateSpecs scans the spec tree rooted at "code-from-spec/",
// validates all nodes against format rules, computes node ranks, and
// checks every declared output artifact for staleness.
//
// It never returns an error. Any problem encountered during scanning or
// validation is recorded in the returned ValidationReport.
func MCPValidateSpecs() *ValidationReport
```

---

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

func main() {
	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, fe := range report.FormatErrors {
		fmt.Printf("format error: node=%s rule=%s detail=%s\n", fe.Node, fe.Rule, fe.Detail)
	}

	for _, name := range report.Cycles {
		fmt.Printf("cycle detected: %s\n", name)
	}

	for _, se := range report.Staleness {
		fmt.Printf(
			"staleness: node=%s output=%s path=%s status=%s rank=%d detail=%s\n",
			se.Node, se.OutputID, se.ArtifactPath, se.Status, se.Rank, se.Detail,
		)
	}

	_ = spectreevalidate.FormatError{}
}
```
