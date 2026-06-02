[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@PgeIEYeEPPH7w5JI_U5zHZLGlD8)

# Package `frontmatter`

```go
package frontmatter
```

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter`

## Types

```go
package frontmatter

type FrontmatterExternal struct {
	Path string
}

type Frontmatter struct {
	DependsOn []string
	External  []*FrontmatterExternal
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

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

func FrontmatterParse(file_path *pathutils.PathCfs) (*Frontmatter, error)
```

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
	path := &pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"}

	fm, err := frontmatter.FrontmatterParse(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("depends_on:", fm.DependsOn)
	fmt.Println("input:", fm.Input)
	fmt.Println("output:", fm.Output)
	for _, ext := range fm.External {
		fmt.Println("external path:", ext.Path)
	}
}
```
