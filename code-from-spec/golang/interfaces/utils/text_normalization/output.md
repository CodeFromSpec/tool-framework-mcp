[//]: # (code-from-spec: ROOT/golang/interfaces/utils/text_normalization@CLPByLzWxcWyzlZ2k5V1SC2hgxs)

# Package `textnormalization`

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
```

## Functions

```go
package textnormalization

// NormalizeText trims leading and trailing whitespace from raw,
// collapses any interior runs of whitespace to a single space,
// and converts all characters to their lowercase equivalents.
// Characters with Unicode folding equivalents (e.g. "ß" → "ss")
// are expanded accordingly. An empty string returns an empty string.
func NormalizeText(raw string) string
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

func main() {
	inputs := []string{
		"  Interface  ",
		"PUBLIC",
		"Straße",
		"Testes   de   aceitação",
		"",
	}

	for _, input := range inputs {
		result := textnormalization.NormalizeText(input)
		fmt.Printf("%q → %q\n", input, result)
	}
	// Output:
	// "  Interface  " → "interface"
	// "PUBLIC" → "public"
	// "Straße" → "strasse"
	// "Testes   de   aceitação" → "testes de aceitação"
	// "" → ""
}
```

## Examples

| Input | Output |
|---|---|
| `"  Interface  "` | `"interface"` |
| `"PUBLIC"` | `"public"` |
| `"Straße"` | `"strasse"` |
| `"Testes   de   aceitação"` | `"testes de aceitação"` |
| `""` | `""` |
