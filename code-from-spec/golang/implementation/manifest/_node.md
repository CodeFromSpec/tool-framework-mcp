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

Implement the manifest component as a Go package.

## Logic

### ManifestOpen

1. If mode is not "read" and not "write", return
   ErrInvalidMode.

2. If mode is "read":
   a. Try OpenFile on "code-from-spec/.manifest" with
      mode "read" and timeout 30000.
      If OpenFile returns FileUnreadable (file does not
      exist): return ManifestHandle with mode "read",
      version "v5", entries as empty map.
      If OpenFile returns any other error, propagate it.
      Let manifest_handle be the result.
   b. Try OpenFile on "code-from-spec/.manifest.lock"
      with mode "read" and timeout 30000.
      If OpenFile returns FileUnreadable (lock file does
      not exist):
        i.  Try OpenFile on
            "code-from-spec/.manifest.lock" with mode
            "append" and timeout 0, then .Close() on
            it. Ignore any errors from these calls.
        ii. Retry OpenFile on
            "code-from-spec/.manifest.lock" with mode
            "read" and timeout 30000.
            If this returns any error, propagate it.
      If OpenFile returns LockTimeout, return
      ErrLockTimeout.
      If OpenFile returns any other error, propagate it.
      Let lock_handle be the result.
   c. Parse manifest_handle line by line into an entries
      map (see parsing steps below).
   d. lock_handle.Close() (releases shared lock).
   e. manifest_handle.Close().
   f. Return ManifestHandle with mode "read", version
      "v5", entries set to the parsed entries map.
      No resources are held after return.

3. If mode is "write":
   a. Let lock_handle be the result of OpenFile on
      "code-from-spec/.manifest.lock" with mode
      "append" and timeout 30000.
      If OpenFile returns LockTimeout, return
      ErrLockTimeout.
      If OpenFile returns any other error, propagate it.
      (Lock file is created if it does not exist;
      exclusive lock is now held.)
   b. Try OpenFile on "code-from-spec/.manifest" with
      mode "read" and timeout 30000.
      If OpenFile returns FileUnreadable (file does not
      exist): let entries be an empty map.
      If OpenFile returns any other error, propagate it.
      Else:
        let manifest_handle be the result.
        Parse manifest_handle line by line into an
        entries map (see parsing steps below).
        manifest_handle.Close().
   c. Return ManifestHandle with mode "write", version
      "v5", entries set to the parsed (or empty) entries
      map, and lock_handle retained internally until
      save or discard.

Parsing steps (shared by read and write paths):
  i.   Read the first line with handle.ReadLine().
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

### ManifestSave

1. If handle.Mode is "read", return ErrWrongMode.
2. If handle is already closed (lockHandle is nil),
   return ErrHandleClosed.
3. Let file_handle be OpenFile on
   "code-from-spec/.manifest" with mode "overwrite"
   and timeout 30000.
   If OpenFile returns any error, propagate it.
4. Write the header line with file_handle.Write():
     "code-from-spec: v5\n"
5. Sort the keys of handle.Entries alphabetically.
6. For each key in sorted order:
     Let entry be handle.Entries[key].
     Write the following line with file_handle.Write():
       "<key>;path:<entry.Path>;checksum:<entry.Checksum>;chain:<entry.ChainHash>\n"
7. file_handle.Close().
8. lockHandle.Close() (releases exclusive lock).
   Set handle.closed = true, handle.lockHandle = nil.

### ManifestDiscard

1. If handle.Mode is "read", return ErrWrongMode.
2. If handle is already closed (lockHandle is nil),
   return ErrHandleClosed.
3. lockHandle.Close() (releases exclusive lock).
   Set handle.closed = true, handle.lockHandle = nil.
   Changes to handle.Entries are abandoned.

## Go-specific guidance

- The package name is `manifest`.
- Use the `oslayer` package for `OpenFile`, `CfsPath`,
  and file methods (`ReadLine`, `Write`, `Close`).
- Use `sort.Strings` for sorting entry keys.
- Use `strings.SplitN` for parsing entry lines.
- Use `strings.TrimPrefix` for removing field prefixes.
- Define sentinel errors: `ErrInvalidMode`,
  `ErrLockTimeout`, `ErrWrongMode`, `ErrHandleClosed`.
- Wrap file errors with `fmt.Errorf` + `%w`.
