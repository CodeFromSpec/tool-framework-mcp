# code-from-spec: SPEC/golang/interfaces/mcp_tools/load_chain@vQdiJGGss3_43ZBXIz0sGrt3ntU

## Package

```go
package mcploadchain
```

## Import Path

```
github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain
```

## Error Sentinels

```go
package mcploadchain

import "errors"

var ErrNoOutput          = errors.New("no output")
var ErrInvalidOutputPath = errors.New("invalid output path")
```

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the chain for the given logical name and returns
// a single string containing the chain hash and all context sections.
//
// The returned string uses the following format:
//
//	chain_hash: <27-character hash>
//	--- context ---
//	<context content>
//	--- input ---
//	<input content>
//	--- existing artifact ---
//	<existing artifact content>
//
// The "--- input ---" section is only present when the target node's
// frontmatter has a non-empty input field.
//
// The "--- existing artifact ---" section is only present when the output
// file exists on disk and is readable.
//
// Errors:
//   - ErrNoOutput: target node has no output field.
//   - ErrInvalidOutputPath: the output path fails path validation.
//   - Propagated from LogicalNameToPath, ChainResolve, ChainHashCompute,
//     NodeParse, and FileOpen.
func MCPLoadChain(logicalName string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain"
)

func main() {
	result, err := mcploadchain.MCPLoadChain("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```
