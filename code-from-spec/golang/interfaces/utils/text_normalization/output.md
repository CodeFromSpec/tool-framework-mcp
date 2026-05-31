[//]: # (code-from-spec: ROOT/golang/interfaces/utils/text_normalization@ReqXVnYDBbG4KdvHzJkzR6yBYp4)

# Package `textnormalization`

**Import path:**

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
```

## Overview

The `textnormalization` package provides utilities for normalizing text strings — trimming whitespace, lowercasing, and expanding Unicode characters such as ligatures and diacritics.

---

## Functions

### `NormalizeText`

```go
// NormalizeText trims leading and trailing whitespace from raw,
// converts it to lowercase, and expands Unicode characters
// (e.g., "Straße" → "strasse"). Multiple internal spaces are
// collapsed to a single space. Returns an empty string unchanged.
func NormalizeText(raw string) string
```

---

## Examples

### Typical call pattern

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

---

## Input / Output Reference

| Input | Output |
|---|---|
| `"  Interface  "` | `"interface"` |
| `"PUBLIC"` | `"public"` |
| `"Straße"` | `"strasse"` |
| `"Testes   de   aceitação"` | `"testes de aceitação"` |
| `""` | `""` |
