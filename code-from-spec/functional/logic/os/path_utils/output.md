<!-- code-from-spec: ROOT/functional/logic/os/path_utils@fYw7SltKhOfCv7bi6MYUZJt5Pm8 -->

namespace: pathutils

---

record PathCfs
  value: string

record PathOs
  value: string

---

function PathGetProjectRoot() -> PathOs

  1. Read the current working directory of the process.
     If the working directory cannot be read, raise error "cannot determine project root".

  2. Return a PathOs with value set to the working directory.

---

function PathValidateCfs(value: string)

  1. If value is empty, raise error "path is empty".

  2. If value starts with "/", raise error "path is absolute".
     If value matches a drive letter pattern (e.g. starts with a letter followed by ":"),
     raise error "path is absolute".

  3. If value contains "\", raise error "path contains backslash".

  4. Normalize the path by resolving "." and ".." components.

  5. For each component in the normalized path,
     if the component is "..", raise error "directory traversal".

---

function PathCfsToOs(cfs_path: PathCfs) -> PathOs

  1. Call PathValidateCfs with cfs_path.value.
     If it raises an error, propagate that error.

  2. Call PathGetProjectRoot() to get the project root.
     If it raises an error, propagate that error.

  3. Replace all forward slashes in cfs_path.value with the OS-native separator.

  4. Join the project root value with the converted path to form an absolute path.

  5. If the resulting path exists on disk,
     resolve symlinks on the resulting path.
     If the resolved path does not start with the project root value,
     raise error "resolves outside root".

  6. Return a PathOs with the absolute path value.

---

function PathOsToCfs(os_path: PathOs) -> PathCfs

  1. Call PathGetProjectRoot() to get the project root.
     If it raises an error, propagate that error.

  2. If os_path.value exists on disk,
     resolve symlinks on os_path.value.
     Use the resolved path for the remaining steps.

  3. If the path (resolved or original) does not start with the project root value,
     raise error "resolves outside root".

  4. Compute the relative path by removing the project root prefix from the path.
     Remove any leading OS separator from the result.

  5. Replace all OS-native separators in the relative path with "/".

  6. Return a PathCfs with the resulting relative value.
