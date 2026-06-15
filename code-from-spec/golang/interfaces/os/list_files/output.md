[//]: # (code-from-spec: SPEC/golang/interfaces/os/list_files@cEs1vrGnFCK3DHuaJ4PC22FtpcY)

# Package `listfiles`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/listfiles`

## Error Sentinels

```go
package listfiles

import "errors"

var ErrDirectoryNotFound = errors.New("directory not found")
var ErrWalkError         = errors.New("filesystem error occurred while traversing")
```

## Functions

```go
package listfiles

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// ListFiles returns all files (not directories) found recursively under the
// given directory. Results are pathutils.PathCfs values sorted alphabetically.
// If the directory exists but contains no files, returns an empty slice.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	dir := &pathutils.PathCfs{Value: "code-from-spec/functional"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Value)
	}
}
```
