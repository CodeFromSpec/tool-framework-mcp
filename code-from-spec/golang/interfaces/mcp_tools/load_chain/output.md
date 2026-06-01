[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@lSs04uHCDrqWP2QWMV5cIEf6QRE)

# Package `mcploadchain`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
```

Package `mcploadchain` implements the `load_chain` MCP tool, which resolves and assembles the spec chain for a target node.

---

## Structs

```go
package mcploadchain

// MCPLoadChainResult holds the result returned by MCPLoadChain.
type MCPLoadChainResult struct {
	ChainHash string
	Context   string
	Input     *string
}
```

---

## Error Sentinels

```go
package mcploadchain

import "errors"

// ErrNoOutputs is returned when the target node has no outputs field.
var ErrNoOutputs = errors.New("no outputs")

// ErrInvalidOutputPath is returned when an output path fails path validation.
var ErrInvalidOutputPath = errors.New("invalid output path")
```

---

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the spec chain for the given logical name and returns
// the assembled context, chain hash, and optional input content.
//
// Errors:
//   - ErrNoOutputs: the target node has no outputs field.
//   - ErrInvalidOutputPath: an output path fails path validation.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (ChainResolver.*): propagated from ChainResolve.
//   - (ChainHash.*): propagated from ChainHashCompute.
//   - (NodeParsing.*): propagated from NodeParse.
//   - (FileReader.*): propagated from FileOpen.
func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
)

func main() {
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
	if err != nil {
		if errors.Is(err, mcploadchain.ErrNoOutputs) {
			log.Fatal("target node has no outputs field")
		}
		if errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
			log.Fatal("an output path failed validation")
		}
		log.Fatalf("load_chain failed: %v", err)
	}

	fmt.Println("chain hash:", result.ChainHash)
	fmt.Println("context length:", len(result.Context))

	if result.Input != nil {
		fmt.Println("input content length:", len(*result.Input))
	}
}
```
