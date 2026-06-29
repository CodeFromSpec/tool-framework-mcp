---
depends_on:
  - ARTIFACT/domain/code-from-spec/manifest-format
  - SPEC/golang/implementation/oslayer(interface)
output: internal/manifest/manifest.go
---

# SPEC/golang/implementation/manifest

Manages the `.manifest` file that tracks the state of
every generated artifact.

# Public

## Package

`package manifest`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"`

## Interface

```go
type ManifestEntry struct {
    Path      string
    Checksum  string
    ChainHash string
}

type Manifest struct {
    Version string
    Entries map[string]ManifestEntry
    // unexported: readOnly, closed, lockHandle
}

func OpenManifest(readOnly bool) (*Manifest, error)
func (m *Manifest) Save() error
func (m *Manifest) Discard() error
```

### OpenManifest

Opens the manifest for reading or writing.

**readOnly = true:** returns a snapshot of the manifest
entries. If the manifest file does not exist, returns
an empty entries map. The caller gets an in-memory
copy — no resources are held after the call returns.
Concurrent readers do not block each other.
`Save` and `Discard` return `ErrReadOnly`.

**readOnly = false:** loads the manifest entries and
holds an exclusive lock until `Save` or `Discard` is
called. If the manifest file does not exist, the
entries map is empty. Only one writer at a time —
concurrent writers block until the lock is released.

Errors: `ErrLockTimeout`, propagated oslayer errors.

### Save

Writes the entries map to disk, creating the manifest
file if it does not exist. Entries are serialized in
alphabetical order by logical name. Releases the lock
and marks the manifest as closed.

Errors: `ErrReadOnly`, `ErrManifestClosed`, propagated
oslayer errors.

### Discard

Releases the lock and marks the manifest as closed
without writing. Changes to the entries map are
discarded.

Errors: `ErrReadOnly`, `ErrManifestClosed`.

## Constraints

- The entries map uses the artifact logical name as key
  (e.g., `ARTIFACT/payments/fees/calculation`).
- Callers operate on `m.Entries` directly — read,
  add, modify, or remove entries.

# Agent

Implement the manifest component as a Go package.

## Logic

### OpenManifest

1. If readOnly is true:
   a. Try oslayer.OpenFile on
      "code-from-spec/.manifest" with mode "read" and
      timeout 30000.
      If OpenFile returns oslayer.ErrFileUnreadable (file
      does not exist): return Manifest with readOnly =
      true, Version = "v5", Entries as empty map.
      If OpenFile returns any other error, propagate it.
      Let manifest_file be the result.
   b. Try oslayer.OpenFile on
      "code-from-spec/.manifest.lock" with mode "read"
      and timeout 30000.
      If OpenFile returns oslayer.ErrFileUnreadable (lock
      file does not exist):
        i.  Try oslayer.OpenFile on
            "code-from-spec/.manifest.lock" with mode
            "append" and timeout 0, then .Close() on
            it. Ignore any errors from these calls.
        ii. Retry oslayer.OpenFile on
            "code-from-spec/.manifest.lock" with mode
            "read" and timeout 30000.
            If this returns any error, propagate it.
      If OpenFile returns oslayer.ErrLockTimeout, return
      ErrLockTimeout.
      If OpenFile returns any other error, propagate it.
      Let lock_file be the result.
   c. Parse manifest_file line by line into an entries
      map (see parsing steps below).
   d. lock_file.Close() (releases shared lock).
   e. manifest_file.Close().
   f. Return Manifest with readOnly = true, Version =
      "v5", Entries set to the parsed entries map.
      No resources are held after return.

2. If readOnly is false:
   a. Let lock_file be the result of oslayer.OpenFile on
      "code-from-spec/.manifest.lock" with mode
      "append" and timeout 30000.
      If OpenFile returns oslayer.ErrLockTimeout, return
      ErrLockTimeout.
      If OpenFile returns any other error, propagate it.
      (Lock file is created if it does not exist;
      exclusive lock is now held.)
   b. Try oslayer.OpenFile on
      "code-from-spec/.manifest" with mode "read" and
      timeout 30000.
      If OpenFile returns oslayer.ErrFileUnreadable (file
      does not exist): let entries be an empty map.
      If OpenFile returns any other error, propagate it.
      Else:
        let manifest_file be the result.
        Parse manifest_file line by line into an
        entries map (see parsing steps below).
        manifest_file.Close().
   c. Return Manifest with readOnly = false, Version =
      "v5", Entries set to the parsed (or empty) entries
      map, and lock_file retained internally until
      Save or Discard.

Parsing steps (shared by read and write paths):
  i.   Read the first line with file.ReadLine().
       If the line is not "code-from-spec: v5", return
       ErrManifestFormatError (unexpected header).
  ii.  For each subsequent line (read until EndOfFile):
       Split the line on ";" into fields.
       If the line has fewer than 4 fields, skip it.
       Let name     be field[0].
       Let path_val be field[1] with the leading "path:"
         prefix removed.
       Let checksum be field[2] with the leading
         "checksum:" prefix removed.
       Let chain    be field[3] with the leading "chain:"
         prefix removed.
       Store ManifestEntry(Path: path_val,
         Checksum: checksum, ChainHash: chain)
       in entries map under key name.

### Save

1. If m.readOnly is true, return ErrReadOnly.
2. If m.closed is true, return ErrManifestClosed.
3. Let file be oslayer.OpenFile on
   "code-from-spec/.manifest" with mode "overwrite"
   and timeout 30000.
   If OpenFile returns any error, propagate it.
4. Write the header line with file.Write():
     "code-from-spec: v5\n"
5. Sort the keys of m.Entries alphabetically.
6. For each key in sorted order:
     Let entry be m.Entries[key].
     Write the following line with file.Write():
       "<key>;path:<entry.Path>;checksum:<entry.Checksum>;chain:<entry.ChainHash>\n"
7. file.Close().
8. m.lockFile.Close() (releases exclusive lock).
   Set m.closed = true.

### Discard

1. If m.readOnly is true, return ErrReadOnly.
2. If m.closed is true, return ErrManifestClosed.
3. m.lockFile.Close() (releases exclusive lock).
   Set m.closed = true.
   Changes to m.Entries are abandoned.

## Go-specific guidance

- The package name is `manifest`.
- Use the `oslayer` package for `OpenFile`, `CfsPath`,
  and file methods (`ReadLine`, `Write`, `Close`).
- Use `sort.Strings` for sorting entry keys.
- Use `strings.SplitN` for parsing entry lines.
- Use `strings.TrimPrefix` for removing field prefixes.
- Define sentinel errors: `ErrLockTimeout`,
  `ErrReadOnly`, `ErrManifestClosed`,
  `ErrManifestFormatError`.
- Wrap file errors with `fmt.Errorf` + `%w`.
