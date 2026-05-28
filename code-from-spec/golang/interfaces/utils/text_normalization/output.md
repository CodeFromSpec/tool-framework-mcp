# Interface: `textnormalization`

<!-- code-from-spec: ROOT/golang/interfaces/utils/text_normalization@GfkyT8qLvX2L836LC7lEGjk1fZs -->

## Package

```go
package textnormalization
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/textnormalization"
```

## Functions

```go
// NormalizeText normalizes a raw string by trimming leading and trailing
// whitespace, collapsing internal runs of whitespace to a single space,
// converting the result to lowercase, and transliterating Unicode characters
// to their ASCII equivalents where possible (e.g., "ß" → "ss").
//
// An empty string input returns an empty string.
func NormalizeText(rawString string) string
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/textnormalization"
)

func main() {
	fmt.Println(textnormalization.NormalizeText("  Interface  "))
	// Output: interface

	fmt.Println(textnormalization.NormalizeText("PUBLIC"))
	// Output: public

	fmt.Println(textnormalization.NormalizeText("Straße"))
	// Output: strasse

	fmt.Println(textnormalization.NormalizeText("Testes   de   aceitação"))
	// Output: testes de aceitação

	fmt.Println(textnormalization.NormalizeText(""))
	// Output:
}
```
