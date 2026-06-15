<!-- code-from-spec: SPEC/functional/logic/os/path_utils@Az6UqJzCibehaKRh46bpmnV18WM -->

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

  2. Return the working directory as a PathOs.

---

function PathValidateCfs(value: string)

  1. If value is empty, raise error "path is empty".

  2. If value starts with "/" or matches a drive letter pattern (e.g. "C:"),
     raise error "path is absolute".

  3. If value contains "\", raise error "path contains backslash".

  4. Normalize the path by resolving "." and ".." components.

  5. For each component in the normalized path,
     if the component is "..", raise error "directory traversal detected".

---

function PathCfsToOs(cfs_path: PathCfs) -> PathOs

  1. Call PathValidateCfs with cfs_path.value.
     If it raises an error, propagate the error.

  2. Call PathGetProjectRoot.
     If it raises an error, propagate the error.
     Store the result as root.

  3. Replace all "/" characters in cfs_path.value with the OS-native path separator.
     Store the result as os_relative.

  4. Join root.value and os_relative to form an absolute path.
     Store the result as absolute_path.

  5. If absolute_path exists on disk,
       resolve symlinks on absolute_path to get resolved_path.
       If resolved_path does not start with root.value,
         raise error "resolves outside root".
       Set absolute_path to resolved_path.

  6. Return absolute_path as a PathOs.

---

function PathOsToCfs(os_path: PathOs) -> PathCfs

  1. Call PathGetProjectRoot.
     If it raises an error, propagate the error.
     Store the result as root.

  2. If os_path.value exists on disk,
       resolve symlinks on os_path.value to get resolved_path.
       Set os_path.value to resolved_path.

  3. If os_path.value does not start with root.value,
     raise error "resolves outside root".

  4. Compute the relative portion of os_path.value by removing the root.value prefix
     and any leading path separator.
     Store the result as relative_path.

  5. Replace all OS-native path separator characters in relative_path with "/".
     Store the result as cfs_value.

  6. Return cfs_value as a PathCfs.
