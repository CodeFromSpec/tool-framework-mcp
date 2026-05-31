<!-- code-from-spec: ROOT/golang/interfaces/mcp_tools/hash_fragment@VBAXbFE-DyX6Dw6ZeGcrI-PeYIM -->

# Interface: `mcphashfragment`

## Package

```go
package mcphashfragment
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcphashfragment"
```

## Error Sentinels

```go
import "errors"

// ErrInvalidLineRange is returned when the lines parameter does not
// match the expected format, when start < 1, when start > end, or
// when end exceeds the total number of lines in the file.
var ErrInvalidLineRange = errors.New("invalid line range")
```

Errors propagated from other packages (not re-declared here):

- `(PathUtils.*)` — propagated from `PathValidateCfs`
- `(FileReader.*)` — propagated from `FileOpen`

## Function Signatures

```go
// MCPHashFragment reads the lines denoted by the range string from the
// file at path (relative to the project root) and returns a SHA-1
// digest of that fragment encoded as a 27-character base64url string
// (RFC 4648 §5, no padding).
//
// path must be a forward-slash relative path validated by PathValidateCfs.
// lines must be a range of the form "start-end" (e.g. "150-210") where
// start >= 1 and end <= the total number of lines in the file.
//
// Errors:
//   - ErrInvalidLineRange: the range format is invalid, start < 1,
//     start > end, or end exceeds the file's line count.
//   - PathUtils errors propagated from PathValidateCfs.
//   - FileReader errors propagated from FileOpen.
func MCPHashFragment(path string, lines string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcphashfragment"
)

func main() {
	// Compute the hash of lines 150 through 210 of a file.
	hash, err := mcphashfragment.MCPHashFragment("internal/chainhash/chainhash.go", "150-210")
	if err != nil {
		log.Fatalf("hash_fragment failed: %v", err)
	}

	// hash is a 27-character base64url string, e.g. "VBAXbFE-DyX6Dw6ZeGcrI-PeYIM"
	fmt.Println(hash)
}
```
