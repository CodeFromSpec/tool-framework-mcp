[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@P3Id2yq3QTDZ6T5JeiLXUOW719I)

# Interface: `listfiles`

**Package:** `package listfiles`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"`

---

## Dependencies

```go
import (
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)
```

---

## Error Sentinels

```go
var (
    // ErrDirectoryNotFound is returned when the given directory does not exist.
    ErrDirectoryNotFound = errors.New("directory not found")

    // ErrWalkError is returned when a filesystem error occurs while traversing
    // the directory tree.
    ErrWalkError = errors.New("walk error")
)
```

---

## Functions

```go
// ListFiles returns all files (not directories) found recursively under
// the given directory. Results are PathCfs values sorted alphabetically.
// If the directory exists but contains no files, an empty slice is returned.
//
// Returns an error if:
//   - validation of cfs_path fails (errors propagated from PathCfsToOs).
//   - conversion of discovered OS paths to CFS paths fails (errors
//     propagated from PathOsToCfs).
//   - the directory does not exist (ErrDirectoryNotFound).
//   - a filesystem error occurs while traversing (ErrWalkError).
func ListFiles(cfs_path *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
    dir := &pathutils.PathCfs{Value: "code-from-spec/functional"}

    files, err := listfiles.ListFiles(dir)
    if err != nil {
        log.Fatalf("could not list files: %v", err)
    }

    for _, f := range files {
        fmt.Println(f.Value)
    }
}
```
