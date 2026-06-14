# code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@1IyOWc2KLmKt9W4CzhveaT3c378

# Package `mcploadchain`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain`

---

## Error Sentinels

```go
package mcploadchain

import "errors"

var ErrNoOutput          = errors.New("target node has no output field")
var ErrInvalidOutputPath = errors.New("the output path fails path validation")
```

---

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the full chain for the given logical name and returns
// a single formatted string containing the chain hash, context content, and
// optionally the input and existing artifact sections.
//
// The returned string has the following structure:
//
//   chain_hash: <27-character hash>
//   --- context ---
//   <context content>
//   --- input ---
//   <input content>
//   --- existing artifact ---
//   <existing artifact content>
//
// The "--- input ---" section is only present when the target node's
// frontmatter has a non-empty input field.
//
// The "--- existing artifact ---" section is only present when the output
// file exists on disk and is readable.
func MCPLoadChain(logical_name string) (string, error)
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
	result, err := mcploadchain.MCPLoadChain("SPEC/payments/invoices")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```
