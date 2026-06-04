<!-- code-from-spec: ROOT/functional/logic/os/list_files@sVEU6SIgoJIbOWDXvKdabXnLSVw -->

function ListFiles(cfs_path: pathutils.PathCfs) -> list of pathutils.PathCfs

  1. Call pathutils.PathCfsToOs(cfs_path) to get an OS path.
     If it raises an error, propagate it.

  2. Check that the resolved OS path exists and is a directory.
     If it does not exist, raise error "DirectoryNotFound".

  3. Recursively walk the directory tree rooted at the OS path.
     If the walk encounters a filesystem error, raise error "WalkError".

  4. For each entry encountered during the walk:
     If the entry is a directory, skip it (continue traversal).
     If the entry is a file, call PathOsToCfs with its absolute OS path.
       If PathOsToCfs raises an error, propagate it.
       Otherwise, add the resulting pathutils.PathCfs to the result list.

  5. Sort the result list alphabetically by the PathCfs value field.

  6. Return the sorted result list.
     If no files were found, return an empty list.
