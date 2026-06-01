[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@wP8p9Z_uscFQJ3hZ29qAbhwEqd8)

# Package `frontmatter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
```

Package `frontmatter` parses YAML frontmatter from spec node files and returns structured metadata.

---

## Structs

```go
package frontmatter

// FrontmatterExternal represents an external file reference declared
// in a node's frontmatter.
type FrontmatterExternal struct {
	Path string
}

// FrontmatterOutput represents a single output entry declared in a
// node's frontmatter.
type FrontmatterOutput struct {
	ID   string
	Path string
}

// Frontmatter holds the parsed metadata extracted from a spec node file.
// All fields default to their zero value when absent from the YAML.
type Frontmatter struct {
	DependsOn []string
	External  []*FrontmatterExternal
	Input     string
	Outputs   []*FrontmatterOutput
}
```

---

## Error Sentinels

```go
package frontmatter

import "errors"

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML")
```

---

## Functions

```go
package frontmatter

// FrontmatterParse reads the file at filePath, extracts the YAML
// frontmatter block delimited by --- markers, and returns the parsed
// Frontmatter. All fields default to empty (empty slice, empty string)
// when absent from the YAML.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrMalformedYAML: the content between --- delimiters is not valid YAML.
//   - (FileReader.*): propagated from FileOpen.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/os/file_reader/_node.md"}

	fm, err := frontmatter.FrontmatterParse(cfsPath)
	if err != nil {
		if errors.Is(err, frontmatter.ErrFileUnreadable) {
			log.Fatal("could not read spec file")
		}
		if errors.Is(err, frontmatter.ErrMalformedYAML) {
			log.Fatal("frontmatter contains invalid YAML")
		}
		log.Fatalf("parse failed: %v", err)
	}

	fmt.Println("input:", fm.Input)
	fmt.Println("depends_on:", fm.DependsOn)

	for _, o := range fm.Outputs {
		fmt.Printf("output: id=%s path=%s\n", o.ID, o.Path)
	}

	for _, e := range fm.External {
		fmt.Printf("external: path=%s\n", e.Path)
	}
}
```
