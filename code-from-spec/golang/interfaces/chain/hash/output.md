[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@Su4iOa8FdoeVVCyENdO-I75Q-wQ)

# Package `chainhash`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
```

Package `chainhash` computes a stable 27-character base64url-encoded SHA-1 hash for a resolved spec chain, used to detect staleness of generated artifacts.

---

## Error Sentinels

```go
package chainhash

import "errors"

// ErrFileUnreadable is returned when a file in the chain cannot be
// read or opened.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrParseFailure is returned when a node file cannot be parsed.
var ErrParseFailure = errors.New("parse failure")
```

---

## Functions

```go
package chainhash

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"

// ChainHashCompute receives a Chain (as returned by ChainResolve) and
// returns a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a
// content hash (SHA-1) for each file, concatenates all content hashes
// as raw bytes in chain assembly order, and computes the final SHA-1
// of the concatenation.
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
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

func main() {
	// First, resolve the chain for a target logical name.
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/hash")
	if err != nil {
		log.Fatalf("failed to resolve chain: %v", err)
	}

	// Compute the chain hash.
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		if errors.Is(err, chainhash.ErrFileUnreadable) {
			log.Fatal("a file in the chain could not be read")
		}
		if errors.Is(err, chainhash.ErrParseFailure) {
			log.Fatal("a node file in the chain could not be parsed")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	// The hash is a 27-character base64url-encoded SHA-1 string.
	fmt.Printf("chain hash: %s\n", hash)
}
```
