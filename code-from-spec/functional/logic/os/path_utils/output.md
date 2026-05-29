<!-- code-from-spec: ROOT/functional/logic/os/path_utils@x0wuKabdM2qeVthJmht4JeoD8IE -->

# path_utils

## Data Structures

```
record PathCfs
  value: string

record PathOs
  value: string
```

---

## function PathGetProjectRoot() -> PathOs

  1. Read the working directory of the current process.
     If the working directory cannot be read, raise error "cannot determine root".

  2. Return the working directory as a PathOs.

---

## function PathValidateCfs(value: string)

  1. If value is empty, raise error "path is empty".

  2. If value starts with "/" or matches the pattern of a drive letter followed
     by ":" (e.g. "C:"), raise error "path is absolute".

  3. If value contains any "\" character, raise error "path contains backslash".

  4. Normalize the path by resolving "." and ".." components using standard
     path normalization rules.

  5. For each component in the normalized path:
     If any component equals "..", raise error "directory traversal".

---

## function PathCfsToOs(cfs_path: PathCfs) -> PathOs

  1. Call PathValidateCfs with cfs_path.value.
     If it raises an error, propagate that error to the caller unchanged.

  2. Replace all "/" characters in cfs_path.value with the OS path separator.

  3. Call PathGetProjectRoot() to obtain the project root as a PathOs.
     If it raises an error, propagate that error to the caller unchanged.

  4. Join the project root path with the converted relative path to form
     a single absolute path.

  5. If the resulting path exists on disk:
     a. Resolve all symlinks in the path to obtain the real absolute path.
     b. If the resolved path does not start with the project root path,
        raise error "resolves outside root".

  6. Return the absolute path as a PathOs.

---

## function PathOsToCfs(os_path: PathOs) -> PathCfs

  1. Call PathGetProjectRoot() to obtain the project root as a PathOs.
     If it raises an error, propagate that error to the caller unchanged.

  2. If os_path.value exists on disk:
     a. Resolve all symlinks in os_path.value to obtain the real absolute path.
     b. Use the resolved path for the remaining steps.

  3. If the path (resolved or original) does not start with the project root path,
     raise error "resolves outside root".

  4. Compute the relative path by removing the project root prefix from the path.
     Strip any leading path separator from the result.

  5. Replace all OS path separator characters with "/".

  6. Return the result as a PathCfs.
