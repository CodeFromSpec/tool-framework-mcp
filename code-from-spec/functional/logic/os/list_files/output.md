<!-- code-from-spec: SPEC/functional/logic/os/list_files@CTNCcxnGTSG-lFm6vUnV5njnY5E -->

function ListFiles(cfs_path: pathutils.PathCfs) -> list of pathutils.PathCfs

  1. Call pathutils.PathCfsToOs(cfs_path) to get an OS path.
     If it raises any error, propagate it to the caller.
     Assign result to os_path.

  2. Check that os_path points to an existing directory.
     If the directory does not exist, raise error "DirectoryNotFound".

  3. Initialize an empty list, results.

  4. Walk the directory at os_path recursively, visiting every entry.
     If the walk itself raises a filesystem error, raise error "WalkError".

     For each entry encountered during the walk:
       If the entry is a directory, skip it (continue traversal but do not add).
       If the entry is a file:
         Call PathOsToCfs(entry_os_path) to convert it to a PathCfs.
         If it raises any error, propagate it to the caller.
         Append the resulting PathCfs to results.

  5. Sort results alphabetically by their value field.

  6. Return results.
