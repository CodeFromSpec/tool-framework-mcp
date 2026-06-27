<!-- code-from-spec: SPEC/functional/tests/manifest@bkZMNhZz7TC2yIqoK8frau8Wp1U -->

## ManifestOpen — read mode — happy path

### Read existing manifest

Setup:
- Create a `.manifest` file with a header line `"v5"` and two entries, each with distinct logical name, path, checksum, and chain_hash fields.

Actions:
1. Call `ManifestOpen("read")`.

Expected outcome:
- Returns a ManifestHandle with `mode` = `"read"`, `version` = `"v5"`, and an entries map containing both entries.
- Each entry has the correct path, checksum, and chain_hash matching the file contents.

---

### Read empty manifest (header only)

Setup:
- Create a `.manifest` file containing only the header line.

Actions:
1. Call `ManifestOpen("read")`.

Expected outcome:
- Returns a ManifestHandle with an empty entries map.

---

### Read missing manifest

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `ManifestOpen("read")`.

Expected outcome:
- Returns a ManifestHandle with an empty entries map.
- No files are created on disk.

---

## ManifestOpen — write mode — happy path

### Write mode loads existing entries

Setup:
- Create a `.manifest` file with a header line and one entry.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard` to release the lock.

Expected outcome:
- `ManifestOpen` returns a ManifestHandle with an entries map containing the one entry.
- `ManifestDiscard` succeeds without error.

---

### Write mode with missing manifest

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard` to release the lock.

Expected outcome:
- `ManifestOpen` returns a ManifestHandle with an empty entries map.
- `ManifestDiscard` succeeds without error.

---

## ManifestSave — happy path

### Save creates manifest from scratch

Setup:
- No `.manifest` file exists on disk.

Actions:
1. Call `ManifestOpen("write")`.
2. Add two entries to the handle's entries map (with distinct logical names, paths, checksums, and chain_hashes).
3. Call `ManifestSave`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains a header line followed by both entries.
- Entries appear in alphabetical order by logical name.

---

### Save overwrites existing manifest

Setup:
- Create a `.manifest` file with one entry (logical name `"alpha"`).

Actions:
1. Call `ManifestOpen("write")`.
2. Add a second entry with logical name `"beta"` to the handle's entries map.
3. Call `ManifestSave`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the header line followed by both entries (`"alpha"` then `"beta"`) in alphabetical order.

---

### Save with modified entry

Setup:
- Create a `.manifest` file with one entry (logical name `"alpha"`, checksum `"old-checksum"`).

Actions:
1. Call `ManifestOpen("write")`.
2. Modify the checksum of the `"alpha"` entry in the handle's entries map to `"new-checksum"`.
3. Call `ManifestSave`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the `"alpha"` entry with checksum `"new-checksum"`.

---

### Save with removed entry

Setup:
- Create a `.manifest` file with two entries (logical names `"alpha"` and `"beta"`).

Actions:
1. Call `ManifestOpen("write")`.
2. Remove the `"beta"` entry from the handle's entries map.
3. Call `ManifestSave`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains the header line followed by only the `"alpha"` entry.

---

## ManifestDiscard — happy path

### Discard does not modify file

Setup:
- Create a `.manifest` file with one entry (logical name `"alpha"`).

Actions:
1. Call `ManifestOpen("write")`.
2. Add a second entry with logical name `"beta"` to the handle's entries map.
3. Call `ManifestDiscard`.
4. Read the `.manifest` file from disk.

Expected outcome:
- The file contains only the original `"alpha"` entry.
- The `"beta"` addition was discarded.

---

## Wrong mode — failure cases

### Save on read handle

Setup:
- None.

Actions:
1. Call `ManifestOpen("read")`.
2. Call `ManifestSave` on the returned handle.

Expected outcome:
- `ManifestSave` raises `WrongMode` error.

---

### Discard on read handle

Setup:
- None.

Actions:
1. Call `ManifestOpen("read")`.
2. Call `ManifestDiscard` on the returned handle.

Expected outcome:
- `ManifestDiscard` raises `WrongMode` error.

---

## Handle closed — failure cases

### Discard after save

Setup:
- None.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestSave`.
3. Call `ManifestDiscard` on the same handle.

Expected outcome:
- `ManifestDiscard` raises `HandleClosed` error.

---

### Save after discard

Setup:
- None.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard`.
3. Call `ManifestSave` on the same handle.

Expected outcome:
- `ManifestSave` raises `HandleClosed` error.

---

### Save after save

Setup:
- None.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestSave`.
3. Call `ManifestSave` again on the same handle.

Expected outcome:
- The second `ManifestSave` raises `HandleClosed` error.

---

### Discard after discard

Setup:
- None.

Actions:
1. Call `ManifestOpen("write")`.
2. Call `ManifestDiscard`.
3. Call `ManifestDiscard` again on the same handle.

Expected outcome:
- The second `ManifestDiscard` raises `HandleClosed` error.

---

## Invalid mode

### ManifestOpen rejects unknown mode

Setup:
- None.

Actions:
1. Call `ManifestOpen("invalid")`.

Expected outcome:
- Raises `InvalidMode` error.

---

## Concurrency

### Concurrent readers do not block

Setup:
- None.

Actions:
1. Call `ManifestOpen("read")` from one execution context.
2. Call `ManifestOpen("read")` from a second concurrent execution context.

Expected outcome:
- Both calls succeed without either blocking the other.

---

### Writer blocks reader

Setup:
- None.

Actions:
1. Call `ManifestOpen("write")` to acquire the exclusive lock.
2. In a separate concurrent execution context, call `ManifestOpen("read")`.
3. Call `ManifestDiscard` on the writer handle.

Expected outcome:
- The `ManifestOpen("read")` call in step 2 blocks until `ManifestDiscard` is called in step 3.
- After the lock is released, the read call returns a ManifestHandle successfully.

---

### Writer blocks writer

Setup:
- None.

Actions:
1. Call `ManifestOpen("write")` to acquire the exclusive lock.
2. In a separate concurrent execution context, call `ManifestOpen("write")`.
3. Call `ManifestSave` or `ManifestDiscard` on the first writer handle.

Expected outcome:
- The second `ManifestOpen("write")` call in step 2 blocks until the first handle is closed in step 3.
- After the lock is released, the second write call returns a ManifestHandle successfully.
