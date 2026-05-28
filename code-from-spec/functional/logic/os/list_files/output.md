<!-- code-from-spec: ROOT/functional/logic/os/list_files@8EtLi-dEo9vbOFMTP0J0XlEPEQ8 -->

# ListFiles

## Function Signatures

```
function ListFiles(cfs_path: PathCfs) -> list of PathCfs
  errors:
    - (validation errors): propagated from PathCfsToOs.
    - (conversion errors): propagated from PathOsToCfs.
    - "directory not found": the directory does not exist.
    - "walk error": a filesystem error occurred while traversing.
```

## Logic

function ListFiles(cfs_path: PathCfs) -> list of PathCfs

  1. Convert cfs_path to an OS path by calling PathCfsToOs(cfs_path).
     If PathCfsToOs raises an error, propagate it to the caller.
     Store the result as os_path.

  2. Check that os_path refers to an existing directory on the filesystem.
     If the directory does not exist, raise error "directory not found".

  3. Create an empty list called results.

  4. Walk the directory at os_path recursively, visiting every entry
     beneath it (including entries inside subdirectories).
     If the walk itself encounters a filesystem error, raise error "walk error".

     For each entry encountered during the walk:
       If the entry is a directory, skip it (continue to the next entry).
       If the entry is a file:
         Convert the entry's OS path to a PathCfs by calling PathOsToCfs(<entry os path>).
         If PathOsToCfs raises an error, propagate it to the caller.
         Append the resulting PathCfs to results.

  5. Sort results alphabetically by their value field.

  6. Return results.
