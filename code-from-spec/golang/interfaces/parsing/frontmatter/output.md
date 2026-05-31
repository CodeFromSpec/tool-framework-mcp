[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@-SrdYls5RVgHCfYLPDmpV-21x2A)

# Package `frontmatter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
```

Parses frontmatter from spec node files, extracting structured metadata such as dependencies, external fragments, input paths, and output declarations.

---

## Structs

```go
package frontmatter

// FrontmatterExternalFragment represents an optional annotated fragment
// of an external file referenced in the frontmatter.
type FrontmatterExternalFragment struct {
	// Description is an optional human-readable label for the fragment.
	Description string

	// Lines is the line range or content selector string.
	Lines string

	// Hash is a content hash for the fragment.
	Hash string
}

// FrontmatterExternal represents a single external file reference,
// optionally broken into named fragments.
type FrontmatterExternal struct {
	// Path is the CFS-format path to the external file.
	Path string

	// Fragments is an optional list of fragments within the external file.
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single declared output in the frontmatter.
type FrontmatterOutput struct {
	// ID is the logical identifier for the output artifact.
	ID string

	// Path is the CFS-format path where the output file should be written.
	Path string
}

// Frontmatter holds the parsed contents of the YAML frontmatter block
// from a spec node file.
type Frontmatter struct {
	// DependsOn is the list of logical names this node depends on.
	DependsOn []string

	// External is the list of external file references.
	External []*FrontmatterExternal

	// Input is the CFS-format path to the input source material file.
	Input string

	// Outputs is the list of declared output artifacts.
	Outputs []*FrontmatterOutput
}
```

---

## Error Sentinels

```go
package frontmatter

import "errors"

// ErrFileUnreadable is returned when the spec file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML")
```

---

## Functions

```go
package frontmatter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FrontmatterParse opens and parses the spec node file at filePath,
// extracting the YAML frontmatter block delimited by --- markers.
//
// All fields default to their zero values (empty list, empty string)
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
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	cfs := &pathutils.PathCfs{Value: "code-from-spec/golang/interfaces/parsing/frontmatter/_node.md"}

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		log.Fatalf("FrontmatterParse: %v", err)
	}

	fmt.Println("depends_on:", fm.DependsOn)
	fmt.Println("input:", fm.Input)

	for _, out := range fm.Outputs {
		fmt.Printf("output — id: %s, path: %s\n", out.ID, out.Path)
	}

	for _, ext := range fm.External {
		fmt.Printf("external — path: %s, fragments: %d\n", ext.Path, len(ext.Fragments))
		for _, frag := range ext.Fragments {
			fmt.Printf("  fragment — description: %q, lines: %s, hash: %s\n",
				frag.Description, frag.Lines, frag.Hash)
		}
	}
}
```
