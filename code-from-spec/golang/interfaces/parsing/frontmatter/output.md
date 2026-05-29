[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/frontmatter@gGV8u0uB_vNbn-5ZQcX0X6izQMo)

# Interface: `frontmatter`

**Package:** `package frontmatter`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"`

---

## Structs

```go
// FrontmatterExternalFragment represents a single fragment entry within
// an external dependency, identified by an optional description, its
// raw content lines, and a hash.
type FrontmatterExternalFragment struct {
    Description string
    Lines       string
    Hash        string
}

// FrontmatterExternal represents an external dependency declared in a
// spec file's frontmatter, consisting of a path and an optional list
// of fragments.
type FrontmatterExternal struct {
    Path      string
    Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single output entry declared in a
// spec file's frontmatter, with an id and a target path.
type FrontmatterOutput struct {
    ID   string
    Path string
}

// Frontmatter holds the parsed contents of a spec file's YAML
// frontmatter block. All fields default to empty (empty list,
// empty string) when absent from the YAML.
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
// at the given CFS path. It returns a populated Frontmatter with all
// declared fields. Fields absent from the YAML default to empty values
// (empty string, empty list).
//
// Returns an error if:
//   - the path is invalid or cannot be resolved (path errors propagated
//     from the underlying file open operation).
//   - the file cannot be opened or read (ErrFileUnreadable).
//   - the content between --- delimiters is not valid YAML (ErrMalformedYAML).
func FrontmatterParse(file_path *pathutils.PathCfs) (*Frontmatter, error)
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
    cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/os/file_reader/_node.md"}

    fm, err := frontmatter.FrontmatterParse(cfsPath)
    if err != nil {
        log.Fatalf("failed to parse frontmatter: %v", err)
    }

    fmt.Println("depends_on:", fm.DependsOn)
    fmt.Println("input:", fm.Input)

    for _, out := range fm.Outputs {
        fmt.Printf("output — id: %s, path: %s\n", out.ID, out.Path)
    }

    for _, ext := range fm.External {
        fmt.Printf("external — path: %s, fragments: %d\n", ext.Path, len(ext.Fragments))
    }
}
```
