---
depends_on:
  - ARTIFACT/golang/interfaces/os/file
output: code-from-spec/golang/interfaces/manifest/output.md
---

# SPEC/golang/interfaces/manifest

Manages the `.manifest` file that tracks the state of
every generated artifact.

# Public

## Package

`package manifest`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/manifest"`

## Interface

```
record ManifestEntry
  path: string
  checksum: string
  chain_hash: string

record ManifestHandle
  mode: string
  version: string
  entries: map of string to ManifestEntry
```

```go
func ManifestOpen(mode string) (*ManifestHandle, error)
func ManifestSave(handle *ManifestHandle) error
func ManifestDiscard(handle *ManifestHandle) error
```

### ManifestOpen

Opens the manifest for reading or writing.

**Read mode:** returns a snapshot of the manifest
entries. If the manifest file does not exist, returns
an empty entries map. The caller gets an in-memory
copy — no resources are held after the call returns.
Concurrent readers do not block each other.
`ManifestSave` and `ManifestDiscard` on a read handle
return `WrongMode`.

**Write mode:** loads the manifest entries and holds
an exclusive lock until `ManifestSave` or
`ManifestDiscard` is called. If the manifest file does
not exist, the entries map is empty. Only one writer
at a time — concurrent writers block until the lock is
released.

Errors: `InvalidMode`, `LockTimeout`, propagated
`File.*` errors.

### ManifestSave

Writes the entries map to disk, creating the manifest
file if it does not exist. Entries are serialized in
alphabetical order by logical name. Releases the lock
and closes the handle.

Errors: `WrongMode`, `HandleClosed`, propagated
`File.*` errors.

### ManifestDiscard

Releases the lock and closes the handle without
writing. Changes to the entries map are discarded.

Errors: `WrongMode`, `HandleClosed`.

## Constraints

- The entries map uses the artifact logical name as key
  (e.g., `ARTIFACT/payments/fees/calculation`).
- Callers operate on `handle.entries` directly — read,
  add, modify, or remove entries.

# Agent

Generate an interface specification document listing
the package, import path, records, and function
signatures.

Use the `file` package's `FileHandle` type for the
internal lock handle. The `ManifestHandle` struct
should hold a `lockHandle *file.FileHandle` field
(unexported) to track the lock.

Map `ManifestEntry` fields to Go types:
- `path` → `Path string`
- `checksum` → `Checksum string`
- `chain_hash` → `ChainHash string`

Map `ManifestHandle` fields to Go types:
- `mode` → `Mode string`
- `version` → `Version string`
- `entries` → `Entries map[string]ManifestEntry`
- `lockHandle` → `lockHandle *file.FileHandle`
  (unexported)
- `closed` → `closed bool` (unexported)
