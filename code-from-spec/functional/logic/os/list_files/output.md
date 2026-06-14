<!-- code-from-spec: ROOT/functional/logic/os/list_files@TRiOzgM4dk8yOMhuDt96ueJ5XHo -->

## Namespace

None — this module does not export records consumed by other modules.

---

## Function

function ListFiles(cfs_path: pathutils.PathCfs) -> list of pathutils.PathCfs

  1. Call pathutils.PathCfsToOs with cfs_path.
     If it raises a PathUtils error, propagate the error.
     Assign the result to os_path.

  2. Check that os_path refers to an existing directory on the filesystem.
     If the directory does not exist, raise error "DirectoryNotFound".

  3. Initialize an empty list called results.

  4. Walk os_path recursively, visiting every entry under it.
     If the walk encounters a filesystem error, raise error "WalkError".

     For each entry encountered during the walk:
       If the entry is a directory, skip it (do not add to results).
       If the entry is a file:
         Call pathutils.PathOsToCfs with the entry's absolute OS path.
         If it raises a PathUtils error, propagate the error.
         Append the returned pathutils.PathCfs value to results.

  5. Sort results alphabetically by their value field.

  6. Return results.

---

## Error conditions

- DirectoryNotFound: raised in step 2 when os_path does not resolve to an
  existing directory.
- WalkError: raised in step 4 when the filesystem walk returns an error for
  any entry.
- PathUtils.*: any error raised by pathutils.PathCfsToOs (step 1) or
  pathutils.PathOsToCfs (step 4) is propagated to the caller unchanged.
