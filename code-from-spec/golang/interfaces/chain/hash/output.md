[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@ix3VSdJyE1UklaD-HSV7to2yXgY)

# Package `chainhash`

```go
package chainhash
```

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash`

## Types

No additional types. Inputs and outputs use primitive Go types and the
`chainresolver.Chain` type from the `chainresolver` package.

## Error Sentinels

```go
package chainhash

import "errors"

var ErrFileUnreadable = errors.New("file unreadable")
var ErrParseFailure   = errors.New("parse failure")
```

## Functions

```go
package chainhash

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"

// ChainHashCompute receives a Chain (as returned by ChainResolve) and returns
// a 27-character base64url encoded SHA-1 hash. The function reads each
// position's content from disk, computes a SHA-1 content hash for each,
// concatenates all content hashes as raw bytes in chain assembly order, and
// computes the final SHA-1 of the concatenation.
//
// Chain assembly order:
//  1. Ancestors (root-first)
//  2. Dependencies (sorted alphabetically by file path, then qualifier)
//  3. External entries (sorted alphabetically by path)
//  4. Target
//  5. Input (if present)
//
// Returns ErrFileUnreadable if a file in the chain cannot be read or opened,
// ErrParseFailure if a node file cannot be parsed, or propagated errors from
// FileReader and NodeParsing packages.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

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
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		log.Fatal(err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("chain hash:", hash)
}
```
