[//]: # (code-from-spec: SPEC/golang/interfaces/manifest@5tU2oIPY8nTpNKsq4nA1n25f8PQ)

# Package `manifest`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/manifest`

## Struct Definitions

```go
package manifest

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
)

// ManifestEntry holds the recorded state for a single artifact.
type ManifestEntry struct {
	Path      string
	Checksum  string
	ChainHash string
}

// ManifestHandle holds the state for an open manifest, including its mode,
// version, entries map, and — for write mode — an exclusive lock handle.
// The caller must call ManifestSave or ManifestDiscard when done.
type ManifestHandle struct {
	Mode       string
	Version    string
	Entries    map[string]ManifestEntry
	lockHandle *file.FileHandle
	closed     bool
}
```

## Error Sentinels

```go
package manifest

import "errors"

var ErrInvalidMode  = errors.New("invalid mode")
var ErrLockTimeout  = errors.New("lock timeout")
var ErrWrongMode    = errors.New("wrong mode")
var ErrHandleClosed = errors.New("handle closed")
```

## Function Signatures

```go
package manifest

// ManifestOpen opens the manifest for reading or writing.
//
// Read mode ("read"): loads a snapshot of the manifest entries into memory
// and returns immediately. If the manifest file does not exist, Entries is
// an empty map. No resources are held after the call returns. Concurrent
// readers do not block each other. ManifestSave and ManifestDiscard on a
// read handle return ErrWrongMode.
//
// Write mode ("write"): loads the manifest entries and holds an exclusive
// lock until ManifestSave or ManifestDiscard is called. If the manifest
// file does not exist, Entries is an empty map. Only one writer at a time
// — concurrent writers block until the lock is released.
//
// Returns ErrInvalidMode if mode is not "read" or "write".
// Returns ErrLockTimeout if the exclusive lock cannot be acquired in time
// (write mode only).
// Errors from the file package are propagated as-is.
func ManifestOpen(mode string) (*ManifestHandle, error)

// ManifestSave writes the Entries map to disk, creating the manifest file
// if it does not exist. Entries are serialized in alphabetical order by
// logical name. Releases the lock and closes the handle.
//
// Returns ErrWrongMode if the handle was opened in read mode.
// Returns ErrHandleClosed if the handle has already been saved or discarded.
// Errors from the file package are propagated as-is.
func ManifestSave(handle *ManifestHandle) error

// ManifestDiscard releases the lock and closes the handle without writing.
// Changes to the Entries map are discarded.
//
// Returns ErrWrongMode if the handle was opened in read mode.
// Returns ErrHandleClosed if the handle has already been saved or discarded.
func ManifestDiscard(handle *ManifestHandle) error
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/manifest"
)

func main() {
	wh, err := manifest.ManifestOpen("write")
	if err != nil {
		log.Fatal(err)
	}

	wh.Entries["ARTIFACT/payments/fees/calculation"] = manifest.ManifestEntry{
		Path:      "generated/payments/fees/calculation.go",
		Checksum:  "abc123",
		ChainHash: "xyz789",
	}

	delete(wh.Entries, "ARTIFACT/payments/fees/old")

	if err := manifest.ManifestSave(wh); err != nil {
		log.Fatal(err)
	}

	rh, err := manifest.ManifestOpen("read")
	if err != nil {
		log.Fatal(err)
	}

	for name, entry := range rh.Entries {
		fmt.Printf("%s -> %s (checksum: %s)\n", name, entry.Path, entry.Checksum)
	}
}
```
