---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/validate
output: code-from-spec/golang/interfaces/mcp_tools/validate_specs/output.md
---

# SPEC/golang/interfaces/mcp_tools/validate_specs

Validates the spec tree for format errors, circular
references, and artifact staleness.

# Public

## Package

`package mcpvalidatespecs`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpvalidatespecs"`

## Interface

```go
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

type ValidationReport struct {
	FormatErrors []spectreevalidate.FormatError
	Cycles       []string
	Staleness    []StalenessEntry
}

func MCPValidateSpecs() ValidationReport
```

No parameters. Scans the entire spec tree starting from
`code-from-spec/`. Always returns a report — never
returns an error. Problems are collected in the report.

`StalenessEntry.Status` is one of:
- `"missing"` — file does not exist.
- `"stale"` — hash mismatch.
- `"malformed tag"` — file exists but has no artifact
  tag or the tag cannot be parsed.

`StalenessEntry.Rank` is the rank from `NodeRankCompute`.

# Agent

Generate an interface specification document listing
the package, import path, struct definitions, and
function signatures.
