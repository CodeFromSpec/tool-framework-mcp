[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@bTUnWpbTD4K-9AaKo6TUA49WdAg)

# Package `frontmatter`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter`

---

## Structs

```go
package frontmatter

// Frontmatter holds the parsed front matter fields from a spec node file.
type Frontmatter struct {
	DependsOn []string
	Input     string
	Output    string
}
```

---

## Error Sentinels

```go
package frontmatter

import "errors"

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrMalformedYAML  = errors.New("content between --- delimiters is not valid YAML")
```

---

## Functions

```go
package frontmatter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FrontmatterParse opens the file at filePath, extracts the YAML block
// delimited by --- markers, and returns the parsed Frontmatter.
// All fields default to their zero values when absent from the YAML.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	path := &pathutils.PathCfs{Value: "SPEC/payments/fees/_node.md"}

	fm, err := frontmatter.FrontmatterParse(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:", fm.Output)
	fmt.Println("Input:", fm.Input)
	fmt.Println("DependsOn:", fm.DependsOn)
}
```
