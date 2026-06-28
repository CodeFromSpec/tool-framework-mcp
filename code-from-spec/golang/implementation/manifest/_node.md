---
depends_on:
  - ARTIFACT/domain/code-from-spec/manifest-format
  - ARTIFACT/golang/interfaces/manifest
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
output: internal/manifest/manifest.go
---

# SPEC/golang/implementation/manifest

# Agent

Implement the manifest component as a Go package.

## Logic

### ManifestOpen

1. If mode is not "read" and not "write", return
   `InvalidMode` error.

2. If mode is "read":
   a. Try FileOpen on "code-from-spec/.manifest" with
      mode "read" and timeout 30000.
      If FileOpen returns FileUnreadable (file does not
      exist): return ManifestHandle with mode "read",
      version "v5", entries as empty map.
      If FileOpen returns any other error, propagate it.
      Let manifest_handle be the result.
   b. Try FileOpen on "code-from-spec/.manifest.lock"
      with mode "read" and timeout 30000.
      If FileOpen returns FileUnreadable (lock file does
      not exist):
        i.  Try FileOpen on
            "code-from-spec/.manifest.lock" with mode
            "append" and timeout 0, then FileClose on
            it. Ignore any errors from these calls.
        ii. Retry FileOpen on
            "code-from-spec/.manifest.lock" with mode
            "read" and timeout 30000.
            If this returns any error, propagate it.
      If FileOpen returns LockTimeout, return
      `LockTimeout` error.
      If FileOpen returns any other error, propagate it.
      Let lock_handle be the result.
   c. Parse manifest_handle line by line into an entries
      map (see parsing steps below).
   d. FileClose lock_handle (releases shared lock).
   e. FileClose manifest_handle.
   f. Return ManifestHandle with mode "read", version
      "v5", entries set to the parsed entries map.
      No resources are held after return.

3. If mode is "write":
   a. Let lock_handle be the result of FileOpen on
      "code-from-spec/.manifest.lock" with mode
      "append" and timeout 30000.
      If FileOpen returns LockTimeout, return
      `LockTimeout` error.
      If FileOpen returns any other error, propagate it.
      (Lock file is created if it does not exist;
      exclusive lock is now held.)
   b. Try FileOpen on "code-from-spec/.manifest" with
      mode "read" and timeout 30000.
      If FileOpen returns FileUnreadable (file does not
      exist): let entries be an empty map.
      If FileOpen returns any other error, propagate it.
      Else:
        let manifest_handle be the result.
        Parse manifest_handle line by line into an
        entries map (see parsing steps below).
        FileClose manifest_handle.
   c. Return ManifestHandle with mode "write", version
      "v5", entries set to the parsed (or empty) entries
      map, and lock_handle retained internally until
      save or discard.

Parsing steps (shared by read and write paths):
  i.   Read the first line with FileReadLine.
       If the line is not "code-from-spec: v5", return
       error "manifest format error: unexpected header".
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

1. If handle.Mode is "read", return `WrongMode` error.
2. If handle is already closed (lockHandle is nil),
   return `HandleClosed` error.
3. Let file_handle be FileOpen on
   "code-from-spec/.manifest" with mode "overwrite"
   and timeout 30000.
   If FileOpen returns any error, propagate it.
4. Write the header line with FileWrite:
     "code-from-spec: v5\n"
5. Sort the keys of handle.Entries alphabetically.
6. For each key in sorted order:
     Let entry be handle.Entries[key].
     Write the following line with FileWrite:
       "<key>;path:<entry.Path>;checksum:<entry.Checksum>;chain:<entry.ChainHash>\n"
7. FileClose file_handle.
8. FileClose lockHandle (releases exclusive lock).
   Set handle.closed = true, handle.lockHandle = nil.

### ManifestDiscard

1. If handle.Mode is "read", return `WrongMode` error.
2. If handle is already closed (lockHandle is nil),
   return `HandleClosed` error.
3. FileClose lockHandle (releases exclusive lock).
   Set handle.closed = true, handle.lockHandle = nil.
   Changes to handle.Entries are abandoned.

## Go-specific guidance

- The package name is `manifest`.
- Use the `file` package for `FileOpen`, `FileReadLine`,
  `FileWrite`, `FileClose`.
- Use the `pathutils` package for `PathCfs`.
- Use `sort.Strings` for sorting entry keys.
- Use `strings.SplitN` for parsing entry lines.
- Use `strings.TrimPrefix` for removing field prefixes.
- Define sentinel errors: `ErrInvalidMode`,
  `ErrLockTimeout`, `ErrWrongMode`, `ErrHandleClosed`.
- Wrap file errors with `fmt.Errorf` + `%w`.
