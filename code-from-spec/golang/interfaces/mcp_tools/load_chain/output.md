[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@wsXp3hy6G2YGL3gKujbUlM28kEA)

# Package `mcploadchain`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
```

Implements the `load_chain` MCP tool, which resolves and assembles the full spec chain context for a given logical name, returning the chain hash, concatenated context, and optional input artifact content.

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

	// Input contains the content of the input artifact, excluding frontmatter.
	// It is nil when the target node has no input field.
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

// MCPLoadChain resolves and assembles the full spec chain for the given
// logical name and returns the chain hash, concatenated context, and
// optional input content.
//
// The function resolves the chain for the target node, computes its hash,
// and concatenates all chain positions into a single context string. If
// the target node declares an input artifact, its content (excluding
// frontmatter) is returned in the Input field of the result.
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
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
	if err != nil {
		// Check for sentinels owned by this package.
		if errors.Is(err, mcploadchain.ErrNoOutputs) {
			log.Fatalf("target node has no outputs: %v", err)
		}
		if errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
			log.Fatalf("invalid output path: %v", err)
		}

		// Check for propagated sentinels from dependency packages.
		if errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
			log.Fatalf("unreadable frontmatter: %v", err)
		}
		if errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
			log.Fatalf("unresolvable artifact: %v", err)
		}
		if errors.Is(err, chainhash.ErrFileUnreadable) {
			log.Fatalf("file unreadable: %v", err)
		}
		if errors.Is(err, chainhash.ErrParseFailure) {
			log.Fatalf("parse failure: %v", err)
		}

		log.Fatalf("MCPLoadChain: %v", err)
	}

	fmt.Printf("chain hash : %s\n", result.ChainHash)
	fmt.Printf("context    : %d bytes\n", len(result.Context))

	if result.Input != nil {
		fmt.Printf("input      : %d bytes\n", len(*result.Input))
	} else {
		fmt.Println("input      : <none>")
	}
}
```
