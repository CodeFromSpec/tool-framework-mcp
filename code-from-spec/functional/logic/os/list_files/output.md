<!-- code-from-spec: ROOT/functional/logic/os/list_files@IwA9T-yBMKXlg7h1hxfknvvfASw -->

function ListFiles(cfs_path: pathutils.PathCfs) -> list of pathutils.PathCfs
  errors:
    - DirectoryNotFound: the directory does not exist.
    - WalkError: a filesystem error occurred while traversing.
    - (PathUtils.*): propagated from pathutils.PathCfsToOs.
    - (PathUtils.*): propagated from PathOsToCfs.

  1. Call pathutils.PathCfsToOs with cfs_path.
     If it returns an error, propagate the error to the caller.
     Let os_path be the resulting PathOs.

  2. Check that os_path refers to an existing directory.
     If it does not exist, raise error "DirectoryNotFound".

  3. Initialize an empty list called results.

  4. Recursively walk the directory tree rooted at os_path.
     For each entry encountered during the walk:
       If a filesystem error occurs while reading a directory
       or entry, raise error "WalkError".
       If the entry is a directory, skip it (do not add to results,
       but continue traversing its contents).
       If the entry is a file:
         Call PathOsToCfs with the entry's absolute OS path.
         If it returns an error, propagate the error to the caller.
         Append the resulting pathutils.PathCfs to results.

  5. Sort results alphabetically by their value field.

  6. Return results.
