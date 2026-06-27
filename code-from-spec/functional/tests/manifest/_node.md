---
depends_on:
  - SPEC/functional/logic/manifest(interface)
output: code-from-spec/functional/tests/manifest/output.md
---

# SPEC/functional/tests/manifest

Test cases for the manifest component.

# Public

## Test cases

### ManifestOpen read — happy path

#### Read existing manifest

Create a `.manifest` file with a header and two entries.
Call `ManifestOpen("read")`. Expect handle with
`version` = `"v5"`, entries map containing both entries
with correct path, checksum, and chain_hash fields.

#### Read empty manifest (header only)

Create a `.manifest` file with only the header line.
Call `ManifestOpen("read")`. Expect handle with empty
entries map.

#### Read missing manifest

Do not create a `.manifest` file. Call
`ManifestOpen("read")`. Expect handle with empty
entries map. No files created on disk.

### ManifestOpen write — happy path

#### Write mode loads existing entries

Create a `.manifest` file with a header and one entry.
Call `ManifestOpen("write")`. Expect handle with entries
map containing the entry. Call `ManifestDiscard` to
release lock.

#### Write mode with missing manifest

Do not create a `.manifest` file. Call
`ManifestOpen("write")`. Expect handle with empty
entries map. Call `ManifestDiscard` to release lock.

### ManifestSave — happy path

#### Save creates manifest from scratch

Call `ManifestOpen("write")`. Add two entries to the
handle's entries map. Call `ManifestSave`. Read the
`.manifest` file from disk. Expect header line followed
by both entries in alphabetical order.

#### Save overwrites existing manifest

Create a `.manifest` file with one entry. Call
`ManifestOpen("write")`. Add a second entry. Call
`ManifestSave`. Read the file. Expect header plus
both entries in alphabetical order.

#### Save with modified entry

Create a `.manifest` with one entry. Call
`ManifestOpen("write")`. Modify the checksum of the
existing entry. Call `ManifestSave`. Read the file.
Expect the updated checksum.

#### Save with removed entry

Create a `.manifest` with two entries. Call
`ManifestOpen("write")`. Remove one entry from the map.
Call `ManifestSave`. Read the file. Expect header plus
only the remaining entry.

### ManifestDiscard — happy path

#### Discard does not modify file

Create a `.manifest` with one entry. Call
`ManifestOpen("write")`. Add a second entry to the map.
Call `ManifestDiscard`. Read the file from disk. Expect
only the original entry — the addition was discarded.

### Wrong mode — failure cases

#### Save on read handle

Call `ManifestOpen("read")`. Call `ManifestSave`.
Expect WrongMode error.

#### Discard on read handle

Call `ManifestOpen("read")`. Call `ManifestDiscard`.
Expect WrongMode error.

### Handle closed — failure cases

#### Discard after save

Call `ManifestOpen("write")`. Call `ManifestSave`. Then
call `ManifestDiscard`. Expect HandleClosed error.

#### Save after discard

Call `ManifestOpen("write")`. Call `ManifestDiscard`. Then
call `ManifestSave`. Expect HandleClosed error.

#### Save after save

Call `ManifestOpen("write")`. Call `ManifestSave`. Then
call `ManifestSave` again. Expect HandleClosed error.

#### Discard after discard

Call `ManifestOpen("write")`. Call `ManifestDiscard`. Then
call `ManifestDiscard` again. Expect HandleClosed error.

### Invalid mode

#### ManifestOpen rejects unknown mode

Call `ManifestOpen("invalid")`. Expect InvalidMode
error.

### Concurrency

#### Concurrent readers do not block

Call `ManifestOpen("read")` twice (simulating two
readers). Expect both to succeed without blocking.

#### Writer blocks reader

Call `ManifestOpen("write")` to hold exclusive lock.
In a separate goroutine, call `ManifestOpen("read")`.
Expect the read to block until `ManifestDiscard` is
called on the writer handle.

#### Writer blocks writer

Call `ManifestOpen("write")` to hold exclusive lock.
In a separate goroutine, call `ManifestOpen("write")`.
Expect the second write to block until `ManifestSave`
or `ManifestDiscard` is called on the first handle.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function names from the interface:
  `ManifestOpen`, `ManifestSave`, `ManifestDiscard`.
- When creating `.manifest` files for test setup, use
  the manifest format from the dependency chain.
