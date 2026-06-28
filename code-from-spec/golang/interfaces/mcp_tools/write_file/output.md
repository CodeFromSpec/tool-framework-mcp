[//]: # (code-from-spec: SPEC/golang/interfaces/mcp_tools/write_file@ftrO8kDf1d06x5aaawJn6y3tYkM)

# Package `mcpwritefile`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpwritefile`

## Error Sentinels

```go
package mcpwritefile

import "errors"

var ErrQualifierNotAllowed   = errors.New("qualifier not allowed")
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput              = errors.New("no output")
var ErrPathNotInOutput       = errors.New("path not in output")
```

## Functions

```go
package mcpwritefile

// MCPWriteFile validates that path is declared in the output field of the node
// identified by logical_name, then writes content to that path.
// Returns "wrote <path>" on success.
func MCPWriteFile(logicalName string, path string, content string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpwritefile"
)

func main() {
	result, err := mcpwritefile.MCPWriteFile(
		"SPEC/payments/fees",
		"code-from-spec/payments/fees/output.go",
		"package fees\n",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
```
