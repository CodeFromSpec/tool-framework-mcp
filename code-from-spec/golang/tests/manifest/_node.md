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

### ManifestOpen — read mode — happy path

#### Read existing manifest

Setup:
- Create a `.manifest` file with a header line
  `"code-from-spec: v5"` and two entries, each with
  distinct logical name, path, checksum, and chain_hash
  fields.

Actions:
1. Call `ManifestOpen("read")`.

Expected outcome:
- Returns a ManifestHandle with `Mode` = `"read"`,
  `Version` = `"v5"`, and an Entries map containing
  both entries.
- Each entry has the correct Path, Checksum, and
  ChainHash matching the file contents.

#### Read empty manifest (header only)

Setup:
- Create a `.manifest` file containing only the header
  line.

Actions:
1. Call `ManifestOpen("read")`.

Expected outcome:
- Returns a ManifestHandle with an empty Entries map.

#### Read missing manifest

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `ManifestOpen("read")`.

Expected outcome:
- Returns a ManifestHandle with an empty Entries map.
- No files are created on disk.

### ManifestOpen — write mode — happy path

#### Write mode loads existing entries

Setup:
- Create a `.manifest` file with a header line and one
  entry.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard` to release the lock.

Expected outcome:
- `ManifestOpen` returns a ManifestHandle with an
  Entries map containing the one entry.
- `ManifestDiscard` succeeds without error.

#### Write mode with missing manifest

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard` to release the lock.

Expected outcome:
- `ManifestOpen` returns a ManifestHandle with an empty
  Entries map.
- `ManifestDiscard` succeeds without error.

### ManifestSave — happy path

#### Save creates manifest from scratch

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `ManifestOpen("write")`.
2. Add two entries to the handle's Entries map (with
   distinct logical names, paths, checksums, and chain
   hashes).
3. Call `ManifestSave`.
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
1. Call `ManifestOpen("write")`.
2. Add a second entry with logical name
   `"ARTIFACT/beta"` to the handle's Entries map.
3. Call `ManifestSave`.
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
1. Call `ManifestOpen("write")`.
2. Modify the Checksum of the `"ARTIFACT/alpha"` entry
   in the handle's Entries map to `"new-checksum"`.
3. Call `ManifestSave`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the `"ARTIFACT/alpha"` entry with
  checksum `"new-checksum"`.

#### Save with removed entry

Setup:
- Create a `.manifest` file with two entries (logical
  names `"ARTIFACT/alpha"` and `"ARTIFACT/beta"`).

Actions:
1. Call `ManifestOpen("write")`.
2. Remove the `"ARTIFACT/beta"` entry from the Entries
   map.
3. Call `ManifestSave`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the header line followed by only
  the `"ARTIFACT/alpha"` entry.

### ManifestDiscard — happy path

#### Discard does not modify file

Setup:
- Create a `.manifest` file with one entry (logical
  name `"ARTIFACT/alpha"`).

Actions:
1. Call `ManifestOpen("write")`.
2. Add a second entry with logical name
   `"ARTIFACT/beta"` to the Entries map.
3. Call `ManifestDiscard`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains only the original `"ARTIFACT/alpha"`
  entry.
- The `"ARTIFACT/beta"` addition was discarded.

### Wrong mode — failure cases

#### Save on read handle

Actions:
1. Call `ManifestOpen("read")`.
2. Call `ManifestSave` on the returned handle.

Expected outcome:
- `ManifestSave` returns `ErrWrongMode`.

#### Discard on read handle

Actions:
1. Call `ManifestOpen("read")`.
2. Call `ManifestDiscard` on the returned handle.

Expected outcome:
- `ManifestDiscard` returns `ErrWrongMode`.

### Handle closed — failure cases

#### Discard after save

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestSave`.
3. Call `ManifestDiscard` on the same handle.

Expected outcome:
- `ManifestDiscard` returns `ErrHandleClosed`.

#### Save after discard

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard`.
3. Call `ManifestSave` on the same handle.

Expected outcome:
- `ManifestSave` returns `ErrHandleClosed`.

#### Save after save

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestSave`.
3. Call `ManifestSave` again on the same handle.

Expected outcome:
- The second `ManifestSave` returns `ErrHandleClosed`.

#### Discard after discard

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard`.
3. Call `ManifestDiscard` again on the same handle.

Expected outcome:
- The second `ManifestDiscard` returns `ErrHandleClosed`.

### Invalid mode

#### ManifestOpen rejects unknown mode

Actions:
1. Call `ManifestOpen("invalid")`.

Expected outcome:
- Returns `ErrInvalidMode`.

### Concurrency

#### Concurrent readers do not block

Actions:
1. Call `ManifestOpen("read")` from one goroutine.
2. Call `ManifestOpen("read")` from a second goroutine.

Expected outcome:
- Both calls succeed without either blocking the other.

#### Writer blocks reader

Actions:
1. Call `ManifestOpen("write")` to acquire the exclusive
   lock.
2. In a separate goroutine, call
   `ManifestOpen("read")`.
3. Call `ManifestDiscard` on the writer handle.

Expected outcome:
- The `ManifestOpen("read")` in step 2 blocks until
  `ManifestDiscard` is called in step 3.
- After the lock is released, the read call returns a
  ManifestHandle successfully.

#### Writer blocks writer

Actions:
1. Call `ManifestOpen("write")` to acquire the exclusive
   lock.
2. In a separate goroutine, call
   `ManifestOpen("write")`.
3. Call `ManifestSave` or `ManifestDiscard` on the first
   writer handle.

Expected outcome:
- The second `ManifestOpen("write")` in step 2 blocks
  until the first handle is closed in step 3.
- After the lock is released, the second write call
  returns a ManifestHandle successfully.

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
