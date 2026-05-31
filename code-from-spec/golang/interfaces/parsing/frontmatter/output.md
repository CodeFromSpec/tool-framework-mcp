[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@kauW8y9pjHx8VK655wGVLIpSVGY)

# Package `frontmatter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
```

Package `frontmatter` parses YAML frontmatter from spec node files, extracting dependency declarations, external fragment references, input paths, and output descriptors.

---

## Structs

```go
package frontmatter

// FrontmatterExternalFragment represents a single fragment entry
// within an external dependency declaration.
type FrontmatterExternalFragment struct {
	// Description is an optional human-readable description of
	// the fragment. Empty string when absent.
	Description string

	// Lines holds the content lines of the fragment.
	Lines string

	// Hash is a content hash for the fragment.
	Hash string
}

// FrontmatterExternal represents an external file referenced in the
// frontmatter, along with its optional fragment list.
type FrontmatterExternal struct {
	// Path is the CFS path to the external file.
	Path string

	// Fragments is the list of fragments within the external file.
	// Empty slice when absent.
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single entry in the outputs list of
// the frontmatter.
type FrontmatterOutput struct {
	// ID is the identifier for this output artifact.
	ID string

	// Path is the CFS path where the output file should be written.
	Path string
}

// Frontmatter holds all parsed fields from a spec node's YAML
// frontmatter block. All fields default to their zero value (empty
// slice or empty string) when absent from the YAML.
type Frontmatter struct {
	// DependsOn is the list of logical names this node depends on.
	DependsOn []string

	// External is the list of external file references.
	External []*FrontmatterExternal

	// Input is the CFS path to the input material for transformation.
	// Empty string when absent.
	Input string

	// Outputs is the list of output descriptors declared by this node.
	Outputs []*FrontmatterOutput
}
```

---

## Error Sentinels

```go
package frontmatter

import "errors"

// ErrFileUnreadable is returned when the file at the given path cannot
// be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between the --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML in frontmatter")
```

---

## Functions

```go
package frontmatter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FrontmatterParse opens the file at filePath, extracts the YAML
// frontmatter delimited by --- lines, and unmarshals it into a
// Frontmatter struct. All missing fields are returned as empty slices
// or empty strings.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrMalformedYAML: the content between --- delimiters is not
//     valid YAML.
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
	// Construct the CFS path to the spec node file.
	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/parsing/frontmatter/_node.md"}

	// Parse the frontmatter from the file.
	fm, err := frontmatter.FrontmatterParse(cfsPath)
	if err != nil {
		if errors.Is(err, frontmatter.ErrFileUnreadable) {
			log.Fatal("could not read spec file")
		}
		if errors.Is(err, frontmatter.ErrMalformedYAML) {
			log.Fatal("frontmatter contains invalid YAML")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	// Inspect the parsed fields.
	fmt.Println("depends_on:", fm.DependsOn)
	fmt.Println("input:", fm.Input)

	for _, ext := range fm.External {
		fmt.Println("external path:", ext.Path)
		for _, frag := range ext.Fragments {
			fmt.Println("  fragment hash:", frag.Hash)
		}
	}

	for _, out := range fm.Outputs {
		fmt.Printf("output id=%s path=%s\n", out.ID, out.Path)
	}
}
```
