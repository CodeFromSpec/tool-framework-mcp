[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@WaNbNSkjwVfgs_MFHD7rUQFT7do)

# Interface: `mcploadchain`

**Package:** `package mcploadchain`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"`

---

## Structs

```go
// MCPLoadChainResult holds the result returned by MCPLoadChain.
type MCPLoadChainResult struct {
    // ChainHash is the 27-character base64url chain hash for the target node.
    ChainHash string

    // Context is all chain content concatenated as a single stream.
    Context string

    // Input is the content of the input artifact (excluding frontmatter),
    // present only when the target node declares an input field.
    Input *string
}
```

---

## Error Sentinels

```go
var (
    // ErrNoOutputs is returned when the target node has no outputs field.
    ErrNoOutputs = errors.New("no outputs")

    // ErrInvalidOutputPath is returned when an output path fails path validation.
    ErrInvalidOutputPath = errors.New("invalid output path")
)
```

> Errors from `LogicalNames`, `ChainResolver`, `ChainHash`, `NodeParsing`, and
> `FileReader` are propagated directly from their respective packages and are not
> re-declared here.

---

## Functions

```go
// MCPLoadChain builds the full chain context for the given logical name and
// returns a result containing the chain hash, concatenated context, and
// optional input content.
//
// The function:
//  1. Converts logical_name to a file path via LogicalNameToPath.
//  2. Resolves the full chain via ChainResolve.
//  3. Computes the chain hash via ChainHashCompute.
//  4. Reads and concatenates all chain files in assembly order to form Context.
//  5. If the target node declares an input field, reads and returns its content
//     (excluding frontmatter) in the Input field.
//  6. Validates that the target node has an outputs field; returns ErrNoOutputs
//     if absent.
//  7. Validates each output path; returns ErrInvalidOutputPath if any path fails
//     validation.
//
// Returns an error if:
//   - logical_name is invalid (LogicalNames errors propagated).
//   - chain resolution fails (ChainResolver errors propagated).
//   - chain hash computation fails (ChainHash errors propagated).
//   - node parsing fails (NodeParsing errors propagated).
//   - any file cannot be opened or read (FileReader errors propagated).
//   - the target node has no outputs field (ErrNoOutputs).
//   - an output path fails validation (ErrInvalidOutputPath).
func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error)
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
)

func main() {
    result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
    if err != nil {
        log.Fatalf("MCPLoadChain failed: %v", err)
    }

    fmt.Println("Chain hash:", result.ChainHash)
    fmt.Println("Context length:", len(result.Context))

    if result.Input != nil {
        fmt.Println("Input content length:", len(*result.Input))
    }
}
```
