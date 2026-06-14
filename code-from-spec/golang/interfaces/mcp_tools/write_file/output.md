[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@uQJIoTDIDIZ5dmSF9s6klbddtrc)

# Package `mcpwritefile`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile`

---

## Error Sentinels

```go
package mcpwritefile

import "errors"

var ErrQualifierNotAllowed    = errors.New("logical name contains a parenthetical qualifier")
var ErrUnreadableFrontmatter  = errors.New("node frontmatter cannot be parsed")
var ErrNoOutput               = errors.New("target node has no output field")
var ErrPathNotInOutput        = errors.New("path is not declared in the node's output")
```

---

## Functions

```go
package mcpwritefile

// MCPWriteFile verifies that path is declared in the output field of the
// node identified by logical_name, then writes content to that path.
// Returns "wrote <path>" on success.
func MCPWriteFile(logical_name string, path string, content string) (string, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
)

func main() {
	result, err := mcpwritefile.MCPWriteFile(
		"SPEC/payments/fees",
		"code-from-spec/golang/payments/fees/output.md",
		"# Fees\n\nGenerated content here.\n",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
```
