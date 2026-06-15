[//]: # (code-from-spec: SPEC/golang/interfaces/utils/text_normalization@i5XyKO6bCaDe3IdHdTEp3i_FEhA)

# Package `textnormalization`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization`

## Overview

Package `textnormalization` provides utilities for normalizing text strings by trimming whitespace, collapsing internal spaces, and converting to lowercase. Unicode characters that have a simple lowercase folding (e.g., `Straße` → `strasse`) are handled via standard library Unicode normalization.

## Function Signatures

```go
package textnormalization

// NormalizeText trims leading and trailing whitespace, collapses all
// internal whitespace sequences to a single space, and converts the
// result to lowercase. Unicode characters with a simple lowercase
// equivalent (e.g., "Straße" → "strasse") are folded accordingly.
// An empty string returns an empty string.
func NormalizeText(rawString string) string
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
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
