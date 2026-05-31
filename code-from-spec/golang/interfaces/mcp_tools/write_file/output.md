[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@A9pRP4QMDN7KWyr6bSdUohAkFk4)

# Interface: `mcpwritefile`

**Package:** `package mcpwritefile`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"`

---

## Error Sentinels

```go
var (
    // ErrUnreadableFrontmatter is returned when the node's frontmatter
    // cannot be parsed.
    ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

    // ErrNoOutputs is returned when the target node has no outputs field.
    ErrNoOutputs = errors.New("no outputs")

    // ErrPathNotInOutputs is returned when the path is not declared in
    // the node's outputs.
    ErrPathNotInOutputs = errors.New("path not in outputs")
)
```

---

## Functions

```go
// MCPWriteFile is the handler for the write_file MCP tool. It validates
// that the given path is declared in the outputs of the node identified
// by logical_name, then writes the content to that path.
//
// Parameters:
//   - logical_name: logical name of the node whose outputs authorize the write.
//   - path: relative file path from project root (forward slashes).
//   - content: complete file content (UTF-8 text).
//
// Returns a success message of the form "wrote <path>" on success.
//
// Returns an error if:
//   - the node's frontmatter cannot be parsed (ErrUnreadableFrontmatter).
//   - the node has no outputs field (ErrNoOutputs).
//   - path is not declared in the node's outputs (ErrPathNotInOutputs).
//   - the logical name cannot be resolved (LogicalNames.* errors propagated).
//   - the path fails CFS validation (PathUtils.* errors propagated).
//   - the file cannot be written (FileWriter.* errors propagated).
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
        log.Fatalf("write_file failed: %v", err)
    }
    fmt.Println(result) // "wrote code-from-spec/golang/interfaces/mcp_tools/write_file/output.md"
}
```
