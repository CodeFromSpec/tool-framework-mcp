[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@gyJ8Pr-rzb2UD35hNMq2_ikyuWM)

# Package `mcpwritefile`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
```

Implements the MCP `write_file` tool handler. Validates the requested write path against the node's declared outputs frontmatter, then writes the file content to disk.

---

## Error Sentinels

```go
package mcpwritefile

import "errors"

// ErrUnreadableFrontmatter is returned when the node's frontmatter cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrNoOutputs is returned when the target node has no outputs field.
var ErrNoOutputs = errors.New("no outputs")

// ErrPathNotInOutputs is returned when the requested path is not declared
// in the node's outputs.
var ErrPathNotInOutputs = errors.New("path not in outputs")
```

---

## Functions

```go
package mcpwritefile

// MCPWriteFile writes content to path after verifying that the path is
// declared in the outputs of the node identified by logical_name.
//
// Steps:
//  1. Resolve logical_name to a spec file path via LogicalNameToPath.
//  2. Parse the frontmatter of that spec file.
//  3. Confirm the outputs field is present and non-empty.
//  4. Confirm path appears in the outputs list.
//  5. Validate path via PathValidateCfs.
//  6. Write content to path via FileWrite.
//
// Returns "wrote <path>" on success.
//
// Errors:
//   - ErrUnreadableFrontmatter: the node's frontmatter cannot be parsed.
//   - ErrNoOutputs: target node has no outputs field.
//   - ErrPathNotInOutputs: path is not declared in the node's outputs.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (PathUtils.*): propagated from PathValidateCfs.
//   - (FileWriter.*): propagated from FileWrite.
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
		"ROOT/golang/interfaces/mcp_tools/write_file",
		"code-from-spec/golang/interfaces/mcp_tools/write_file/output.md",
		"# Generated content\n",
	)
	if err != nil {
		log.Fatalf("MCPWriteFile: %v", err)
	}
	fmt.Println(result) // wrote code-from-spec/golang/interfaces/mcp_tools/write_file/output.md
}
```
