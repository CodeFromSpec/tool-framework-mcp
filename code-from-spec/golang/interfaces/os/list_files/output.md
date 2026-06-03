[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@tOSkxnqkn2xs472iabKReoSJgOc)

# Package `listfiles`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
```

## Error Sentinels

```go
package listfiles

import "errors"

var ErrDirectoryNotFound = errors.New("directory not found")
var ErrWalkError = errors.New("filesystem walk error")
```

## Functions

```go
package listfiles

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ListFiles returns all files (not directories) found recursively under
// the given directory, as PathCfs values sorted alphabetically.
// Returns an empty list if the directory exists but contains no files.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

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
	dir := &pathutils.PathCfs{Value: "internal/listfiles"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Value)
	}
}
```
