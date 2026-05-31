[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@8g5WFebw3KHbBxTV4CvH4EmdVm4)

# Package `mcploadchain`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
```

Implements the `load_chain` MCP tool. Given a logical name, it resolves the full spec chain, computes the chain hash, and returns the assembled context and optional input artifact.

---

## Structs

```go
package mcploadchain

// MCPLoadChainResult holds the output of a successful MCPLoadChain call.
type MCPLoadChainResult struct {
	// ChainHash is the 27-character base64url chain hash.
	ChainHash string

	// Context contains all chain content concatenated as a single stream.
	Context string

	// Input contains the content of the input artifact (excluding frontmatter),
	// if one exists. Nil when no input artifact is present.
	Input *string
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

// MCPLoadChain resolves the spec chain for the given logical name, computes
// the chain hash, and returns the assembled context along with an optional
// input artifact.
//
// The target node must declare an outputs field. Each output path is validated
// before the result is returned.
//
// Errors:
//   - ErrNoOutputs: target node has no outputs field.
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

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
)

func main() {
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
	if err != nil {
		// Check for sentinel errors specific to this package.
		if errors.Is(err, mcploadchain.ErrNoOutputs) {
			log.Fatalf("target node has no outputs: %v", err)
		}
		if errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
			log.Fatalf("invalid output path: %v", err)
		}

		// Propagated errors from dependency packages can also be inspected.
		if errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
			log.Fatalf("frontmatter unreadable: %v", err)
		}
		if errors.Is(err, chainhash.ErrFileUnreadable) {
			log.Fatalf("file unreadable: %v", err)
		}

		log.Fatalf("MCPLoadChain: %v", err)
	}

	fmt.Printf("chain hash: %s\n", result.ChainHash)
	fmt.Printf("context length: %d bytes\n", len(result.Context))

	if result.Input != nil {
		fmt.Printf("input length: %d bytes\n", len(*result.Input))
	} else {
		fmt.Println("no input artifact")
	}
}
```
