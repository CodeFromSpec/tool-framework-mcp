[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/validate_specs@91qzNWnW0fngfRTmB5OYNBhjqzs)

# Interface: `mcpvalidatespecs`

**Package:** `package mcpvalidatespecs`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"`

---

## Structs

```go
// StalenessEntry describes a single output file that is missing, stale,
// or has a malformed artifact tag. Files whose hash matches are not included.
// Status is one of "missing", "stale", or "malformed tag".
// Rank is the value from NodeRankCompute — entries with equal rank have no
// dependency between them and can be processed in parallel.
type StalenessEntry struct {
    Node         string
    OutputID     string
    ArtifactPath string
    Status       string
    Detail       string
    Rank         int
}

// ValidationReport is the result returned by MCPValidateSpecs. It collects
// all format errors found in the spec tree, logical names involved in
// dependency cycles, and staleness information for generated output files.
type ValidationReport struct {
    FormatErrors []*spectreevalidate.FormatError
    Cycles       []string
    Staleness    []*StalenessEntry
}
```

---

## Functions

```go
// MCPValidateSpecs scans the entire spec tree rooted at "code-from-spec/",
// validates every node against the spec tree format rules, checks for
// dependency cycles via NodeRankCompute, and inspects each declared output
// file for staleness. It always returns a report — it never returns an error.
// All problems are collected inside the report.
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

    if len(report.FormatErrors) == 0 && len(report.Cycles) == 0 && len(report.Staleness) == 0 {
        fmt.Println("spec tree is valid and all outputs are up to date")
        return
    }

    for _, e := range report.FormatErrors {
        fmt.Printf("format error | node: %s | rule: %s | detail: %s\n",
            e.Node, e.Rule, e.Detail)
    }

    for _, name := range report.Cycles {
        fmt.Printf("cycle detected | node: %s\n", name)
    }

    for _, s := range report.Staleness {
        fmt.Printf("staleness | node: %s | output: %s | path: %s | status: %s | rank: %d | detail: %s\n",
            s.Node, s.OutputID, s.ArtifactPath, s.Status, s.Rank, s.Detail)
    }
}
```

> **Note on imports in the usage example:** `spectreevalidate` is imported
> because `ValidationReport.FormatErrors` uses `*spectreevalidate.FormatError`.
> Any caller that accesses that field must import
> `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate`.
