[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@BeLpRyDCMCEcfswHVaYPm6Qrn9k)

# Package `frontmatter`

```go
package frontmatter
```

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter`

## Types

```go
package frontmatter

// FrontmatterExternal holds a reference to an external path dependency.
type FrontmatterExternal struct {
	Path string
}

// Frontmatter contains the parsed frontmatter fields from a spec node file.
// All fields default to their zero value (empty list, empty string) when absent.
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

// FrontmatterParse reads the file at file_path, extracts the YAML frontmatter
// delimited by --- markers, and returns a populated Frontmatter. All fields
// default to empty when absent from the YAML. Returns ErrFileUnreadable if the
// file cannot be opened or read, ErrMalformedYAML if the content between the
// --- delimiters is not valid YAML, or a FileReader error propagated from the
// underlying file open operation.
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
