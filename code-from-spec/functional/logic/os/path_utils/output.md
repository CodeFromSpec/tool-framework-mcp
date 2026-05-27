<!-- code-from-spec: ROOT/functional/logic/os/path_utils@JfImHL6sWEjss2CS_jM87D4eQt8 -->

## Records

```
record PathCfs
  value: string

record PathOs
  value: string
```

---

## function PathGetProjectRoot() -> PathOs

Returns the project root as an absolute OS path, determined
from the working directory of the running process.

1. Read the current working directory of the process.
   If the working directory cannot be read, raise error
   "cannot determine root".

2. Return the working directory as a PathOs.

---

## function PathValidateCfs(value: string)

Validates that a string conforms to the PathCfs format.
Raises an error describing the violation if not.
Does not verify that the file exists or resolve symlinks.

1. If value is empty, raise error "path is empty".

2. If value starts with "/" or matches a drive-letter pattern
   (e.g. "C:"), raise error "path is absolute".

3. If value contains "\" (backslash), raise error
   "path contains backslash".

4. Normalize the path by resolving "." and ".." components.

5. If any component of the normalized path is "..",
   raise error "directory traversal".

---

## function PathCfsToOs(cfs_path: PathCfs) -> PathOs

Validates a PathCfs and converts it to an absolute PathOs.
Never sanitizes — rejects invalid paths by raising an error.
Never creates or modifies files.

1. Call PathValidateCfs with cfs_path.value.
   If it raises an error, propagate that error unchanged.

2. Replace all forward-slash "/" separators in cfs_path.value
   with the OS-native path separator.

3. Call PathGetProjectRoot to obtain the project root.
   If it raises an error, propagate that error.

4. Join the project root with the converted path to form
   an absolute OS path.

5. If the resulting path exists on disk:
   a. Resolve all symlinks in the path.
   b. Resolve all symlinks in the project root.
   c. If the resolved path does not start with the
      resolved project root, raise error "resolves outside root".

6. Return the absolute path as a PathOs.

---

## function PathOsToCfs(os_path: PathOs) -> PathCfs

Converts an absolute PathOs to a PathCfs relative to the
project root.
Never creates or modifies files.

1. If os_path.value exists on disk, resolve all symlinks
   in os_path.value.

2. Call PathGetProjectRoot to obtain the project root.
   If it raises an error, propagate that error.
   If the project root exists on disk, resolve all symlinks
   in the project root as well.

3. If the (resolved) path does not start with the
   (resolved) project root, raise error "resolves outside root".

4. Compute the relative path by removing the project root
   prefix from the path, including any leading separator.

5. Replace all OS-native path separators in the relative path
   with forward slashes "/".

6. Return the result as a PathCfs.
