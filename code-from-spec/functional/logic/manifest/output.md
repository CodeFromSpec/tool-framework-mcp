<!-- code-from-spec: SPEC/functional/logic/manifest@2XycSV1gl2txXgJeCEbSZ72LBOw -->

namespace: manifest

---

record ManifestEntry
  path: string
  checksum: string
  chain_hash: string

record ManifestHandle
  mode: string
  version: string
  entries: map of string to ManifestEntry

---

function ManifestOpen(mode: string) -> ManifestHandle
  errors:
    - InvalidMode: mode is not "read" or "write".
    - LockTimeout: the manifest lock could not be acquired within the timeout.
    - (File.*): propagated from FileOpen.

  1. If mode is not "read" and not "write", raise error "InvalidMode".

  2. If mode is "read":
     a. Try FileOpen on "code-from-spec/.manifest" with mode "read" and timeout 30000.
        If FileOpen raises FileUnreadable (file does not exist):
          return ManifestHandle with mode "read", version "v5", entries as empty map.
        If FileOpen raises any other error, propagate it.
        Let manifest_handle be the result.
     b. Try FileOpen on "code-from-spec/.manifest.lock" with mode "read"
        and timeout 30000.
        If FileOpen raises FileUnreadable (lock file does not exist):
          i.  Try FileOpen on "code-from-spec/.manifest.lock" with mode "append"
              and timeout 0, then FileClose on it. Ignore any errors from these calls.
          ii. Retry FileOpen on "code-from-spec/.manifest.lock" with mode "read"
              and timeout 30000.
              If this raises any error, propagate it.
        If FileOpen raises LockTimeout, raise error "LockTimeout".
        If FileOpen raises any other error, propagate it.
        Let lock_handle be the result.
     c. Parse manifest_handle line by line into an entries map
        (see parsing steps below).
     d. FileClose lock_handle (releases shared lock).
     e. FileClose manifest_handle.
     f. Return ManifestHandle with mode "read", version "v5",
        entries set to the parsed entries map.
        No resources are held after return.

  3. If mode is "write":
     a. Let lock_handle be the result of FileOpen on "code-from-spec/.manifest.lock"
        with mode "append" and timeout 30000.
        If FileOpen raises LockTimeout, raise error "LockTimeout".
        If FileOpen raises any other error, propagate it.
        (Lock file is created if it does not exist; exclusive lock is now held.)
     b. Try FileOpen on "code-from-spec/.manifest" with mode "read" and timeout 30000.
        If FileOpen raises FileUnreadable (file does not exist):
          let entries be an empty map.
        If FileOpen raises any other error, propagate it.
        Else:
          let manifest_handle be the result.
          Parse manifest_handle line by line into an entries map
          (see parsing steps below).
          FileClose manifest_handle.
     c. Return ManifestHandle with mode "write", version "v5",
        entries set to the parsed (or empty) entries map,
        and lock_handle retained internally until save or discard.

  Parsing steps (shared by read and write paths):
    i.   Read the first line with FileReadLine.
         If the line is not "code-from-spec: v5", raise error
         "manifest format error: unexpected header".
    ii.  For each subsequent line (read until EndOfFile):
         Split the line on ";" into fields.
         If the line has fewer than 4 fields, skip it.
         Let name     be field[0] (e.g. "ARTIFACT/payments/fees/calculation").
         Let path_val be field[1] with the leading "path:" prefix removed.
         Let checksum be field[2] with the leading "checksum:" prefix removed.
         Let chain    be field[3] with the leading "chain:" prefix removed.
         Store ManifestEntry(path: path_val, checksum: checksum, chain_hash: chain)
         in entries map under key name.

---

function ManifestSave(handle: ManifestHandle)
  errors:
    - WrongMode: handle was opened in "read" mode.
    - HandleClosed: handle was already saved or discarded.
    - (File.*): propagated from FileOpen, FileWrite.

  1. If handle.mode is "read", raise error "WrongMode".

  2. If handle is already closed (lock_handle is absent), raise error "HandleClosed".

  3. Let file_handle be FileOpen on "code-from-spec/.manifest" with mode "overwrite"
     and timeout 30000.
     If FileOpen raises any error, propagate it.

  4. Write the header line with FileWrite:
       "code-from-spec: v5\n"

  5. Sort the keys of handle.entries alphabetically.

  6. For each key in sorted order:
       Let entry be handle.entries[key].
       Write the following line with FileWrite:
         "<key>;path:<entry.path>;checksum:<entry.checksum>;chain:<entry.chain_hash>\n"

  7. FileClose file_handle.

  8. FileClose lock_handle (releases exclusive lock).
     Mark handle as closed (lock_handle absent).

---

function ManifestDiscard(handle: ManifestHandle)
  errors:
    - WrongMode: handle was opened in "read" mode.
    - HandleClosed: handle was already saved or discarded.

  1. If handle.mode is "read", raise error "WrongMode".

  2. If handle is already closed (lock_handle is absent), raise error "HandleClosed".

  3. FileClose lock_handle (releases exclusive lock).
     Mark handle as closed (lock_handle absent).
     Changes to handle.entries are abandoned.
