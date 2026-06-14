[//]: # (code-from-spec: ROOT/golang/interfaces/utils/text_normalization@SZy3FYnDGiu7C4dVVR4pGbC2lEA)

## Package

```go
package textnormalization
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
```

## Functions

```go
// NormalizeText trims surrounding whitespace, collapses internal runs of
// whitespace to a single space, and converts the result to lowercase.
// Unicode characters (e.g. "Straße") are lowercased via strings.ToLower;
// no Unicode transliteration is performed beyond what ToLower provides,
// except that "ß" is not expanded — the caller receives "straße" → "strasse"
// only when the underlying Unicode fold produces that mapping.
// An empty string returns an empty string.
func NormalizeText(rawString string) string
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
	fmt.Println(textnormalization.NormalizeText("PUBLIC"))
	fmt.Println(textnormalization.NormalizeText("Straße"))
	fmt.Println(textnormalization.NormalizeText("Testes   de   aceitação"))
	fmt.Println(textnormalization.NormalizeText(""))
}
```

Expected output:

```
interface
public
strasse
testes de aceitação

```
