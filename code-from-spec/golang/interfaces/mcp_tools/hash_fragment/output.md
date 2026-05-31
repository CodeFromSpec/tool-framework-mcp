[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/hash_fragment@s4O45OTH0zIyRUdZF90uCYNhb18)

# Interface: `mcphashfragment`

## Package declaration

```go
package mcphashfragment
```

## Import path

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcphashfragment"
```

## Error sentinels

```go
import "errors"

// ErrInvalidLineRange is returned when the lines parameter has an invalid
// format, start < 1, start > end, or end exceeds the file's line count.
var ErrInvalidLineRange = errors.New("invalid line range")
```

> Errors propagated from `PathUtils` and `FileReader` are owned by their
> respective packages and are not re-declared here.

## Function signatures

```go
// MCPHashFragment computes a SHA-1 digest of the specified line range within
// the given file. The digest is base64url encoded (RFC 4648 §5, no padding),
// producing a 27-character string.
//
// path must be a file path relative to the project root using forward slashes.
// lines must be a line range in the form "start-end" (e.g., "150-210").
//
// Errors:
//   - ErrInvalidLineRange: the range format is invalid, start < 1,
//     start > end, or end exceeds the file's line count.
//   - PathUtils errors: propagated from PathValidateCfs.
//   - FileReader errors: propagated from FileOpen.
func MCPHashFragment(path string, lines string) (string, error)
```

## Usage example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcphashfragment"
)

func main() {
	// Compute the SHA-1 hash (base64url, no padding) of lines 150–210
	// in the given source file.
	hash, err := mcphashfragment.MCPHashFragment("internal/mypackage/myfile.go", "150-210")
	if err != nil {
		log.Fatalf("MCPHashFragment failed: %v", err)
	}

	// hash is a 27-character base64url string, e.g. "s4O45OTH0zIyRUdZF90uCYNhb18"
	fmt.Println(hash)
}
```
