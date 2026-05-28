[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@RDH1H45XklKqQkP_otSJqkT5oQo)

# Interface: `frontmatter`

## Package

```go
package frontmatter
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
```

---

## Struct Definitions

```go
// FrontmatterExternalFragment represents a single fragment of an external
// dependency, with an optional description, the raw lines content, and a
// hash identifying the fragment.
type FrontmatterExternalFragment struct {
	Description string
	Lines       string
	Hash        string
}

// FrontmatterExternal represents an external dependency referenced in a
// spec file, identified by its path and containing zero or more fragments.
type FrontmatterExternal struct {
	Path      string
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single output entry declared in a spec
// file's frontmatter, with an id and a target path.
type FrontmatterOutput struct {
	ID   string
	Path string
}

// Frontmatter holds all structured data parsed from the YAML frontmatter
// block of a spec file. All fields default to their zero value (empty
// string, empty slice) when absent from the YAML.
type Frontmatter struct {
	DependsOn []*string
	External  []*FrontmatterExternal
	Input     string
	Outputs   []*FrontmatterOutput
}
```

---

## Error Sentinels

```go
var (
	// ErrFileUnreadable is returned when the file cannot be opened or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrMalformedYAML is returned when the content between --- delimiters
	// is not valid YAML.
	ErrMalformedYAML = errors.New("malformed YAML")
)
```

---

## Functions

```go
// FrontmatterParse opens and parses the YAML frontmatter of the spec file
// at the given CFS path, returning a populated Frontmatter. All fields
// default to empty when absent from the YAML.
//
// Possible errors:
//   - Path errors propagated from FileOpen (ErrPathEmpty, ErrPathAbsolute,
//     ErrPathContainsBackslash, ErrDirectoryTraversal, ErrResolvesOutsideRoot,
//     ErrCannotDetermineRoot)
//   - ErrFileUnreadable
//   - ErrMalformedYAML
func FrontmatterParse(file_path *pathutils.PathCfs) (*Frontmatter, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	cfs := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/parsing/frontmatter/_node.md"}

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Input:", fm.Input)

	for _, out := range fm.Outputs {
		fmt.Printf("Output id=%s path=%s\n", out.ID, out.Path)
	}

	for _, dep := range fm.DependsOn {
		fmt.Println("Depends on:", *dep)
	}

	for _, ext := range fm.External {
		fmt.Printf("External path=%s fragments=%d\n", ext.Path, len(ext.Fragments))
	}
}
```
