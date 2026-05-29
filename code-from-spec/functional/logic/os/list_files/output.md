<!-- code-from-spec: ROOT/functional/logic/os/list_files@8EtLi-dEo9vbOFMTP0J0XlEPEQ8 -->

## ListFiles

```
function ListFiles(cfs_path: PathCfs) -> list of PathCfs
  errors:
    - (validation errors): propagated from PathCfsToOs.
    - (conversion errors): propagated from PathOsToCfs.
    - directory not found: the directory does not exist.
    - walk error: a filesystem error occurred while traversing.
```

### Logic

1. Call PathCfsToOs with cfs_path.
   If it returns an error, propagate that error to the caller.
   Store the result as os_path.

2. Check that the directory at os_path exists on the filesystem.
   If it does not exist, raise error "directory not found".

3. Initialize an empty list called results.

4. Walk the directory at os_path recursively.
   If the walk cannot be started, raise error "walk error".

   For each entry encountered during the walk:
     If the entry is a directory, skip it (do not add to results).
     If the entry is a file:
       Call PathOsToCfs with the entry's absolute OS path.
       If it returns an error, raise error "walk error".
       Append the resulting PathCfs to results.
     If a filesystem error occurs while reading the entry,
       raise error "walk error".

5. Sort results alphabetically by their value field.

6. Return results.
