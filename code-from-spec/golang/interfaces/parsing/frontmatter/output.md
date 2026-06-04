[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@QLwobF5jj6bQdL_90NtWu4rMNfw)

# Package `frontmatter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
```

## Structs

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
var ErrMalformedYAML = errors.New("malformed YAML")
```

## Functions

```go
package frontmatter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FrontmatterParse reads the file at file_path, extracts the YAML block
// between the first pair of --- delimiters, and unmarshals it into a
// Frontmatter. All fields default to their zero values when absent from
// the YAML.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error)
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
	path := &pathutils.PathCfs{Value: "code-from-spec/some/node/_node.md"}

	fm, err := frontmatter.FrontmatterParse(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:", fm.Output)
	fmt.Println("Input:", fm.Input)
	fmt.Println("DependsOn:", fm.DependsOn)
	for _, ext := range fm.External {
		fmt.Println("External path:", ext.Path)
	}
}
```
