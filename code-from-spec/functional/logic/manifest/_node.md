---
depends_on:
  - ARTIFACT/domain/code-from-spec/manifest-format
  - SPEC/functional/logic/os/file(interface)
  - SPEC/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/manifest/output.md
---

# SPEC/functional/logic/manifest

Manages the `.manifest` file that tracks the state of
every generated artifact.

# Public

## Namespace

    namespace: manifest

## Interface

```
record ManifestHandle
  mode: string
  version: string
  entries: map of string to ManifestEntry

record ManifestEntry
  path: string
  checksum: string
  chain_hash: string

function ManifestOpen(mode: string) -> ManifestHandle
  errors:
    - InvalidMode: mode is not "read" or "write".
    - LockTimeout: the manifest lock could not be
      acquired within the timeout.
    - (File.*): propagated from FileOpen.

function ManifestSave(handle: ManifestHandle)
  errors:
    - WrongMode: handle was opened in "read" mode.
    - HandleClosed: handle was already saved or discarded.
    - (File.*): propagated from FileOpen, FileWrite.

function ManifestDiscard(handle: ManifestHandle)
  errors:
    - WrongMode: handle was opened in "read" mode.
    - HandleClosed: handle was already saved or discarded.
```

### ManifestOpen

Opens the manifest for reading or writing.

**Read mode:** returns a snapshot of the manifest
entries. If the manifest file does not exist, returns
an empty entries map. The caller gets an in-memory
copy — no resources are held after the call returns.
Concurrent readers do not block each other.
`ManifestSave` and `ManifestDiscard` on a read handle
raise `WrongMode`.

**Write mode:** loads the manifest entries and holds
an exclusive lock until `ManifestSave` or
`ManifestDiscard` is called. If the manifest file does
not exist, the entries map is empty. Only one writer
at a time — concurrent writers block until the lock is
released.

### ManifestSave

Writes the entries map to disk, creating the manifest
file if it does not exist. Entries are serialized in
alphabetical order by logical name. Releases the lock
and closes the handle.

Fails with `WrongMode` if the handle was opened in
read mode.

### ManifestDiscard

Releases the lock and closes the handle without
writing. Changes to the entries map are discarded.

Fails with `WrongMode` if the handle was opened in
read mode.

## Constraints

- The entries map uses the artifact logical name as key
  (e.g., `ARTIFACT/payments/fees/calculation`).
- Callers operate on `handle.entries` directly — read,
  add, modify, or remove entries.

# Agent

Generate pseudocode for each function in the interface.

## Implementation guidance

- The manifest file is at `code-from-spec/.manifest`.
- Use a dedicated lock file at
  `code-from-spec/.manifest.lock` for concurrency
  control. The lock file is never written to — it is
  only opened for locking.
- Use `FileOpen`, `FileReadLine`, `FileWrite`,
  `FileClose` from the file component for all I/O.
- Parse and write following the manifest format
  provided via the dependency chain.

### Read mode

1. If `.manifest` does not exist, return immediately
   with empty entries. Do not create any files.
2. If `.manifest` exists:
   a. Try `FileOpen` on the lock file with mode `"read"`
      for shared lock.
   b. If the lock file does not exist: try `FileOpen`
      with mode `"append"` then `FileClose`
      (best-effort, ignore errors). Then retry
      `FileOpen` with mode `"read"`.
   c. Read and parse `.manifest` into entries map.
   d. Close the lock file (release shared lock).
   e. Return handle with snapshot. No resource held.

### Write mode

1. `FileOpen` on the lock file with mode `"append"` —
   creates the lock file if it does not exist, acquires
   exclusive lock. Lock held until save or discard.
2. Try to read `.manifest` — if it does not exist,
   entries map is empty.
3. Return handle with lock held.

### ManifestSave

1. Use `FileOpen` with mode `"overwrite"` to write
   `.manifest`.
2. Close the lock file handle (release exclusive lock).

### ManifestDiscard

1. Close the lock file handle (release exclusive lock).
