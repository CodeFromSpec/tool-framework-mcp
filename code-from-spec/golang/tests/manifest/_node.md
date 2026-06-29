---
depends_on:
  - ARTIFACT/domain/code-from-spec/manifest-format
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/manifest
output: internal/manifest/manifest_test.go
---

# SPEC/golang/tests/manifest

# Agent

## Test cases

### OpenManifest — readOnly — happy path

#### Read existing manifest

Setup:
- Create a `.manifest` file with a header line
  `"code-from-spec: v5"` and two entries, each with
  distinct logical name, path, checksum, and chain_hash
  fields.

Actions:
1. Call `manifest.OpenManifest(true)`.

Expected outcome:
- Returns a Manifest with `Version` = `"v5"`, and an
  Entries map containing both entries.
- Each entry has the correct Path, Checksum, and
  ChainHash matching the file contents.

#### Read empty manifest (header only)

Setup:
- Create a `.manifest` file containing only the header
  line.

Actions:
1. Call `manifest.OpenManifest(true)`.

Expected outcome:
- Returns a Manifest with an empty Entries map.

#### Read missing manifest

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `manifest.OpenManifest(true)`.

Expected outcome:
- Returns a Manifest with an empty Entries map.
- No files are created on disk.

### OpenManifest — writable — happy path

#### Write mode loads existing entries

Setup:
- Create a `.manifest` file with a header line and one
  entry.

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Call `m.Discard()` to release the lock.

Expected outcome:
- `OpenManifest` returns a Manifest with an Entries map
  containing the one entry.
- `m.Discard()` succeeds without error.

#### Write mode with missing manifest

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Call `m.Discard()` to release the lock.

Expected outcome:
- `OpenManifest` returns a Manifest with an empty
  Entries map.
- `m.Discard()` succeeds without error.

### Save — happy path

#### Save creates manifest from scratch

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Add two entries to m.Entries (with distinct logical
   names, paths, checksums, and chain hashes).
3. Call `m.Save()`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains a header line followed by both
  entries.
- Entries appear in alphabetical order by logical name.

#### Save overwrites existing manifest

Setup:
- Create a `.manifest` file with one entry (logical
  name `"ARTIFACT/alpha"`).

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Add a second entry with logical name
   `"ARTIFACT/beta"` to m.Entries.
3. Call `m.Save()`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the header line followed by both
  entries (`"ARTIFACT/alpha"` then `"ARTIFACT/beta"`)
  in alphabetical order.

#### Save with modified entry

Setup:
- Create a `.manifest` file with one entry (logical
  name `"ARTIFACT/alpha"`, checksum `"old-checksum"`).

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Modify the Checksum of the `"ARTIFACT/alpha"` entry
   in m.Entries to `"new-checksum"`.
3. Call `m.Save()`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the `"ARTIFACT/alpha"` entry with
  checksum `"new-checksum"`.

#### Save with removed entry

Setup:
- Create a `.manifest` file with two entries (logical
  names `"ARTIFACT/alpha"` and `"ARTIFACT/beta"`).

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Remove the `"ARTIFACT/beta"` entry from m.Entries.
3. Call `m.Save()`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the header line followed by only
  the `"ARTIFACT/alpha"` entry.

#### Save with empty entries

Setup:
- Create a `.manifest` file with two entries.

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Clear `m.Entries` (remove all entries).
3. Call `m.Save()`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains only the header line
  `"code-from-spec: v5"`, no entry lines.

### Discard — happy path

#### Discard does not modify file

Setup:
- Create a `.manifest` file with one entry (logical
  name `"ARTIFACT/alpha"`).

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Add a second entry with logical name
   `"ARTIFACT/beta"` to m.Entries.
3. Call `m.Discard()`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains only the original `"ARTIFACT/alpha"`
  entry.
- The `"ARTIFACT/beta"` addition was discarded.

### OpenManifest — failure cases

#### Invalid header

Setup:
- Create a `.manifest` file whose first line is
  `"invalid-header"`.

Actions:
1. Call `manifest.OpenManifest(true)`.

Expected outcome:
- Returns `manifest.ErrManifestFormatError`.

### ReadOnly — failure cases

#### Save on readOnly manifest

Actions:
1. Call `manifest.OpenManifest(true)`.
2. Call `m.Save()`.

Expected outcome:
- `m.Save()` returns `manifest.ErrReadOnly`.

#### Discard on readOnly manifest

Actions:
1. Call `manifest.OpenManifest(true)`.
2. Call `m.Discard()`.

Expected outcome:
- `m.Discard()` returns `manifest.ErrReadOnly`.

### Closed — failure cases

#### Discard after Save

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Call `m.Save()`.
3. Call `m.Discard()`.

Expected outcome:
- `m.Discard()` returns `manifest.ErrManifestClosed`.

#### Save after Discard

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Call `m.Discard()`.
3. Call `m.Save()`.

Expected outcome:
- `m.Save()` returns `manifest.ErrManifestClosed`.

#### Save after Save

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Call `m.Save()`.
3. Call `m.Save()` again.

Expected outcome:
- The second `m.Save()` returns `manifest.ErrManifestClosed`.

#### Discard after Discard

Actions:
1. Call `manifest.OpenManifest(false)`.
2. Call `m.Discard()`.
3. Call `m.Discard()` again.

Expected outcome:
- The second `m.Discard()` returns `manifest.ErrManifestClosed`.

### Concurrency

#### Concurrent readers do not block

Actions:
1. Call `manifest.OpenManifest(true)` from one goroutine.
2. Call `manifest.OpenManifest(true)` from a second
   goroutine.

Expected outcome:
- Both calls succeed without either blocking the other.

#### Writer blocks reader

Actions:
1. Call `manifest.OpenManifest(false)` to acquire the
   exclusive lock.
2. In a separate goroutine, call
   `manifest.OpenManifest(true)`.
3. Call `m.Discard()` on the writer.

Expected outcome:
- The `OpenManifest(true)` in step 2 blocks until
  `m.Discard()` is called in step 3.
- After the lock is released, the read call returns a
  Manifest successfully.

#### Writer blocks writer

Actions:
1. Call `manifest.OpenManifest(false)` to acquire the
   exclusive lock.
2. In a separate goroutine, call
   `manifest.OpenManifest(false)`.
3. Call `m.Save()` or `m.Discard()` on the first writer.

Expected outcome:
- The second `OpenManifest(false)` in step 2 blocks
  until the first manifest is closed in step 3.
- After the lock is released, the second write call
  returns a Manifest successfully.

## Go-specific guidance

- The package name is `manifest_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory
  to the temp dir, since manifest paths are relative to
  the working directory.
- Create `code-from-spec/` subdirectory in the temp dir
  before each test that needs manifest files.
- Helper to write `.manifest` files: write header
  `"code-from-spec: v5\n"` followed by entry lines in
  the expected format.
- Helper to read `.manifest` files back for assertions.
- For concurrency tests, use goroutines with
  `sync.WaitGroup` and channels for synchronization.
  Use short timeouts to avoid hanging tests.
