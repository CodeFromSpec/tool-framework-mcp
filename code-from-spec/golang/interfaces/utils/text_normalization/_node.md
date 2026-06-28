---
output: code-from-spec/golang/interfaces/utils/text_normalization/output.md
---

# SPEC/golang/interfaces/utils/text_normalization

Normalizes text for comparison.

# Public

## Package

`package textnormalization`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"`

## Interface

```go
func NormalizeText(rawString string) string
```

Pure function. Trims leading/trailing whitespace,
collapses internal whitespace runs to a single space,
applies Unicode simple case folding.

### Examples

| Input | Output |
|---|---|
| `"  Interface  "` | `"interface"` |
| `"PUBLIC"` | `"public"` |
| `"Straße"` | `"strasse"` |
| `""` | `""` |

# Agent

Generate an interface specification document listing
the package, import path, and function signature.
