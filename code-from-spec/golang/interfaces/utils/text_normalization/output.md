[//]: # (code-from-spec: ROOT/golang/interfaces/utils/text_normalization@2u9bAusD_uVqBCUrABlRJKwDfQo)

# Package `textnormalization`

```go
package textnormalization
```

Import path:

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
```

## Functions

```go
// NormalizeText trims leading and trailing whitespace from raw_string,
// collapses internal runs of whitespace to a single space, and converts
// the result to lower-case. Special Unicode characters such as the German
// sharp-s (ß) are folded to their ASCII equivalents where a standard
// lower-case mapping exists (e.g. "Straße" → "strasse").
func NormalizeText(raw_string string) string
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

func main() {
	fmt.Println(textnormalization.NormalizeText("  Interface  "))       // "interface"
	fmt.Println(textnormalization.NormalizeText("PUBLIC"))              // "public"
	fmt.Println(textnormalization.NormalizeText("Straße"))              // "strasse"
	fmt.Println(textnormalization.NormalizeText("Testes   de   aceitação")) // "testes de aceitação"
	fmt.Println(textnormalization.NormalizeText(""))                    // ""
}
```
