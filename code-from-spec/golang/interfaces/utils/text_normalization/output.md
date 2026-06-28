[//]: # (code-from-spec: SPEC/golang/interfaces/utils/text_normalization@sZiAaAEnZy799dbIViyeTRFe7UI)

# Package `textnormalization`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization`

## Overview

Package `textnormalization` provides utilities for normalizing text strings by trimming whitespace, collapsing internal spaces, and applying Unicode simple case folding.

## Function Signatures

```go
package textnormalization

// NormalizeText trims leading and trailing whitespace, collapses all
// internal whitespace sequences to a single space, and applies Unicode
// simple case folding (e.g., "Straße" → "strasse").
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
	fmt.Println(textnormalization.NormalizeText(""))
}
```

Expected output:

```
interface
public
strasse

```
