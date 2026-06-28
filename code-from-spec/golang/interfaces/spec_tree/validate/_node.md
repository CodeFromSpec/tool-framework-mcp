---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
output: code-from-spec/golang/interfaces/spec_tree/validate/output.md
---

# SPEC/golang/interfaces/spec_tree/validate

Linter for the spec tree. Receives discovered nodes with
their parsed frontmatter and body, checks structural
rules, and reports all violations found.

# Public

## Package

`package spectreevalidate`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectreevalidate"`

## Interface

```go
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func SpecTreeValidate(entries []*SpecTreeValidateInput, allDirs []string) []*FormatError
```

Takes the full set of discovered nodes with their parsed
frontmatter and body, plus a list of all subdirectory
paths found under `code-from-spec/`. Returns a list of
format errors (empty if all nodes are valid).

# Agent

Generate an interface specification document listing
the package, import path, struct definitions, and
function signatures.
