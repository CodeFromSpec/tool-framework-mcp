[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@SoSZFF7wevgfJYi8sc6jCWiLBpw)

# Package `chainhash`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
```

Computes a deterministic hash over an ordered chain of spec nodes.

---

## Error Sentinels

```go
package chainhash

import "errors"

// ErrFileUnreadable is returned when a file in the chain cannot be read or opened.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrParseFailure is returned when a node file cannot be parsed.
var ErrParseFailure = errors.New("parse failure")
```

---

## Functions

```go
package chainhash

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"

// ChainHashCompute receives a Chain (as returned by ChainResolve) and returns
// a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a content
// hash (SHA-1) for each, concatenates all content hashes as raw bytes in
// chain assembly order, and computes the final SHA-1 of the concatenation.
//
// Errors:
//   - ErrFileUnreadable: a file in the chain cannot be read or opened.
//   - ErrParseFailure: a node file cannot be parsed.
//   - (FileReader.*): propagated from FileOpen.
//   - (NodeParsing.*): propagated from NodeParse.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/hash")
	if err != nil {
		log.Fatalf("ChainResolve: %v", err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		log.Fatalf("ChainHashCompute: %v", err)
	}

	fmt.Printf("chain hash: %s\n", hash)

	// Sentinel errors can be checked with errors.Is:
	//
	//   _, err := chainhash.ChainHashCompute(chain)
	//   if errors.Is(err, chainhash.ErrFileUnreadable) { ... }
	//   if errors.Is(err, chainhash.ErrParseFailure) { ... }
}
```
