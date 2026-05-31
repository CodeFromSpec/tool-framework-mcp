<!-- code-from-spec: ROOT/functional/logic/os/list_files@xFehnKk2CjUoZgl0ztYRC0ZQI-Q -->

function ListFiles(cfs_path: PathCfs) -> list of PathCfs

  1. Call PathCfsToOs(cfs_path) to get an OS-native absolute path.
     If PathCfsToOs returns any error, propagate it to the caller.

  2. Check that the path refers to an existing directory.
     If the directory does not exist, raise error "DirectoryNotFound".

  3. Walk the directory tree recursively starting at the OS path from step 1.
     If the filesystem walk cannot be initiated or encounters a fatal error,
     raise error "WalkError".

  4. For each entry encountered during the walk:
     If the entry is a directory, skip it (do not include, but continue traversal).
     If the entry is a file:
       a. Call PathOsToCfs(entry_os_path) to convert the file's OS path to a PathCfs.
          If PathOsToCfs returns any error, propagate it to the caller.
       b. Append the resulting PathCfs to the result list.

  5. Sort the result list alphabetically by the PathCfs value field.

  6. Return the sorted list.
     If no files were found, return an empty list.
