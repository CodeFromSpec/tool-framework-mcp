[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@jRHEJCgbtqVbpgpmT9TOX3-TmFc)

# Interface: `chainhash`

## Package

```go
package chainhash
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
```

---

## Error Sentinels

```go
var (
	// ErrUnreadableFile is returned when a file in the chain cannot be
	// read or opened.
	ErrUnreadableFile = errors.New("file unreadable")

	// ErrParseFailure is returned when a node file cannot be parsed.
	ErrParseFailure = errors.New("parse failure")
)
```

---

## Functions

```go
// ChainHashCompute receives a Chain (as returned by ChainResolve) and
// returns a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a
// content hash (SHA-1) for each, concatenates all content hashes as
// raw bytes in chain assembly order, and computes the final SHA-1 of
// the concatenation.
//
// Chain assembly order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — from the target's depends_on, sorted alphabetically
//     by file path then by qualifier.
//  3. External — from the target's external, sorted alphabetically by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
//
// Possible errors:
//   - ErrUnreadableFile — a file in the chain cannot be read or opened.
//   - ErrParseFailure — a node file cannot be parsed.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

---

## Usage Examples

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
		log.Fatal(err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hash:", hash) // e.g. "jRHEJCgbtqVbpgpmT9TOX3-TmFc"
}
```
