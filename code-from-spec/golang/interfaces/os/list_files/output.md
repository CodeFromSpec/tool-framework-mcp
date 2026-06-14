[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@Z4_KAUWbMw5h6C30IA8ePrGRdPM)

# Package `listfiles`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles`

---

## Error Sentinels

```go
package listfiles

import "errors"

var ErrDirectoryNotFound = errors.New("directory does not exist")
var ErrWalkError         = errors.New("filesystem error while traversing directory")
```

---

## Functions

```go
package listfiles

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ListFiles returns all files (not directories) found recursively under
// the given directory, as PathCfs values sorted alphabetically.
// Returns an empty slice if the directory exists but contains no files.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
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
	dir := &pathutils.PathCfs{Value: "code-from-spec/SPEC/payments"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Value)
	}
}
```
