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

Returns the current working directory of the process as an absolute PathOs.

1. Read the working directory of the running process.
   If the working directory cannot be read, raise error "cannot determine root".

2. Return a PathOs with that absolute directory path as its value.

---

## function PathValidateCfs(value: string)

Validates that a string conforms to the PathCfs format rules.
Raises a descriptive error on the first violation found.

1. If value is empty, raise error "path is empty".

2. If value starts with "/" or starts with a drive letter pattern
   (one letter followed by ":"), raise error "path is absolute".

3. If value contains any "\" character, raise error "path contains backslash".

4. Normalize the path by resolving any "." and ".." components
   using standard path normalization rules (without touching the filesystem).

5. Split the normalized path into its individual components using "/"
   as the separator.
   For each component:
     If the component equals "..", raise error "directory traversal".

---

## function PathCfsToOs(cfs_path: PathCfs) -> PathOs

Validates a PathCfs and converts it to an absolute PathOs.
Does not require the target path to exist on disk.
Never sanitizes input — rejects any invalid path.

1. Call PathValidateCfs with cfs_path.value.
   If PathValidateCfs raises an error, propagate that error immediately.

2. Replace every "/" character in cfs_path.value with the OS-native
   path separator.

3. Call PathGetProjectRoot.
   If PathGetProjectRoot raises an error, propagate that error.

4. Join the project root path with the converted path from step 2
   to form a single absolute path.

5. If the joined path exists on disk:
     Resolve all symlinks in the joined path to obtain its real absolute path.
     If the real absolute path does not start with the project root path
     (followed by a separator or being exactly the root), raise error
     "resolves outside root".

6. Return a PathOs with the absolute joined path (pre-symlink-resolution)
   as its value.
   Note: if symlinks were resolved in step 5 and containment was confirmed,
   the returned value is still the joined path from step 4, not the
   symlink-resolved path — the symlink check is a security gate only.

---

## function PathOsToCfs(os_path: PathOs) -> PathCfs

Converts an absolute PathOs to a PathCfs relative to the project root.
Does not require the target path to exist on disk.
Never sanitizes input — rejects any path outside the project root.

1. Call PathGetProjectRoot.
   If PathGetProjectRoot raises an error, propagate that error.

2. If os_path.value exists on disk:
     Resolve all symlinks in os_path.value to obtain its real absolute path.
     Use the resolved path for all subsequent steps.
   Else:
     Use os_path.value as-is for all subsequent steps.

3. Verify that the path from step 2 starts with the project root path
   (followed by a separator or being exactly the root).
   If it does not, raise error "resolves outside root".

4. Compute the relative portion of the path by removing the project root
   prefix (and any trailing separator that follows it) from the path.

5. Replace every OS-native path separator in the relative portion with "/".

6. Return a PathCfs with the resulting relative string as its value.
```
