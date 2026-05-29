[//]: # (code-from-spec: ROOT/golang/interfaces/utils/text_normalization@KGEXuPdpOD_H_aqrR9oZrO6lzsM)

# Interface: `textnormalization`

## Package

```go
package textnormalization
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
```

## Function Signatures

```go
// NormalizeText trims leading and trailing whitespace from raw_string,
// collapses internal runs of whitespace to a single space, and converts
// the result to lowercase. Returns an empty string when raw_string is empty
// or contains only whitespace.
func NormalizeText(raw_string string) string
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
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
