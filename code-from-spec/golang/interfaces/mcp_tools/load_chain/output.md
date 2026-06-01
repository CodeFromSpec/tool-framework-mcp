[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@y6eQCq8oQixNQQd-yj4O0V9DYpM)

# Package `mcploadchain`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
```

Package `mcploadchain` implements the `load_chain` MCP tool, which resolves a spec chain for a given logical name, computes its hash, and returns the concatenated chain context together with an optional input artifact.

---

## Structs

```go
package mcploadchain

// MCPLoadChainResult holds the result of a successful load_chain tool call.
type MCPLoadChainResult struct {
	// ChainHash is the 27-character base64url-encoded SHA-1 chain hash.
	ChainHash string

	// Context is all chain content concatenated as a single stream.
	Context string

	// Input is the content of the input artifact, excluding frontmatter.
	// It is nil when the target node has no input artifact.
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

// ErrInvalidOutputPath is returned when an output path fails path
// validation.
var ErrInvalidOutputPath = errors.New("invalid output path")
```

---

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the spec chain for the given logical name,
// computes the chain hash, and returns the concatenated context and
// optional input artifact content.
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

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
)

func main() {
	// Call MCPLoadChain with the logical name of the target node.
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
	if err != nil {
		if errors.Is(err, mcploadchain.ErrNoOutputs) {
			log.Fatal("target node has no outputs field")
		}
		if errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
			log.Fatal("an output path failed validation")
		}
		if errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
			log.Fatal("could not parse a node's frontmatter")
		}
		if errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
			log.Fatal("an ARTIFACT/ reference could not be resolved")
		}
		if errors.Is(err, chainhash.ErrFileUnreadable) {
			log.Fatal("a file in the chain could not be read")
		}
		if errors.Is(err, chainhash.ErrParseFailure) {
			log.Fatal("a node file in the chain could not be parsed")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	// The chain hash is a 27-character base64url-encoded SHA-1 string.
	fmt.Printf("chain_hash: %s\n", result.ChainHash)

	// The context contains all chain content concatenated.
	fmt.Printf("context length: %d bytes\n", len(result.Context))

	// Input is only present when the target node declares an input artifact.
	if result.Input != nil {
		fmt.Printf("input length: %d bytes\n", len(*result.Input))
	} else {
		fmt.Println("input: (none)")
	}
}
```
