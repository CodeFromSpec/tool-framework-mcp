[//]: # (code-from-spec: SPEC/golang/interfaces/parsing/frontmatter@HmYWX-Mr57AB9wDZtUdKfYDOv5M)

# Package `frontmatter`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter`

## Types

```go
package frontmatter

type Frontmatter struct {
	DependsOn []string
	Input     string
	Output    string
}
```

## Error Sentinels

```go
package frontmatter

import "errors"

var ErrFileUnreadable = errors.New("file unreadable")
var ErrMalformedYAML  = errors.New("malformed YAML")
```

## Functions

```go
package frontmatter

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FrontmatterParse opens the file at filePath, extracts the YAML front matter
// delimited by "---" markers, and returns the parsed Frontmatter.
// All fields default to their zero value (empty slice, empty string) when
// absent from the YAML block.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	path := &pathutils.PathCfs{Value: "SPEC/myproject/logic/rules.md"}

	fm, err := frontmatter.FrontmatterParse(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:", fm.Output)
	fmt.Println("Input:", fm.Input)
	fmt.Println("DependsOn:", fm.DependsOn)
}
```
