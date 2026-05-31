[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/hash_fragment@ywMOxq2ekK_UzD1oiY73bi7wP3M)

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

// ErrInvalidLineRange is returned when the line range format is invalid,
// the start line is less than 1, start exceeds end, or end exceeds the
// file's total line count.
var ErrInvalidLineRange = errors.New("invalid line range")
```

## Function Signatures

```go
// MCPHashFragment reads lines [start, end] (inclusive, 1-based) from the
// file at path (relative to the project root, forward slashes), computes
// a SHA-1 digest of those lines, and returns it as a base64url-encoded
// string (RFC 4648 §5, no padding, 27 characters).
//
// path must pass PathUtils.PathValidateCfs validation.
// lines must be a range string of the form "start-end" (e.g. "150-210").
//
// Errors:
//   - ErrInvalidLineRange: range format is invalid, start < 1,
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
	// Compute the SHA-1 hash of lines 150 through 210 of a source file.
	hash, err := mcphashfragment.MCPHashFragment("internal/mypackage/myfile.go", "150-210")
	if err != nil {
		log.Fatalf("hash_fragment failed: %v", err)
	}

	// hash is a 27-character base64url string (RFC 4648 §5, no padding).
	fmt.Println(hash)
}
```
