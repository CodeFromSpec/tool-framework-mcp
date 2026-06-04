<!-- code-from-spec: ROOT/functional/logic/os/path_utils@_VnP8BeAxjhz_-jLTkQrlenvq-4 -->

namespace: pathutils

---

record PathCfs
  value: string

record PathOs
  value: string

---

function PathGetProjectRoot() -> PathOs
  errors:
    - CannotDetermineRoot

  1. Read the current working directory of the process.
     If the working directory cannot be read, raise error "cannot determine project root".

  2. Return a PathOs with the working directory as its value.

---

function PathValidateCfs(value: string)
  errors:
    - PathEmpty
    - PathAbsolute
    - PathContainsBackslash
    - DirectoryTraversal

  1. If value is empty, raise error "path is empty".

  2. If value starts with "/" or starts with a drive letter pattern
     (a single letter followed by ":"), raise error "path is absolute".

  3. If value contains any "\" character, raise error "path contains backslash".

  4. Normalize the path by resolving "." and ".." components.

  5. For each component in the normalized path:
       If the component is "..", raise error "directory traversal".

---

function PathCfsToOs(cfs_path: PathCfs) -> PathOs
  errors:
    - ResolvesOutsideRoot
    - PathEmpty          (propagated from PathValidateCfs)
    - PathAbsolute       (propagated from PathValidateCfs)
    - PathContainsBackslash (propagated from PathValidateCfs)
    - DirectoryTraversal (propagated from PathValidateCfs)
    - CannotDetermineRoot (propagated from PathGetProjectRoot)

  1. Call PathValidateCfs with cfs_path.value.
     If it raises any error, propagate that error immediately without continuing.

  2. Call PathGetProjectRoot.
     If it raises an error, propagate that error immediately without continuing.
     Store the result as root.

  3. Replace all "/" characters in cfs_path.value with the OS-native path separator.
     Store the result as os_relative.

  4. Join root.value and os_relative to form an absolute path.
     Store the result as abs_path.

  5. If abs_path refers to an existing filesystem entry:
       Resolve any symlinks in abs_path to obtain resolved_path.
       If resolved_path does not start with root.value, raise error "resolves outside root".
       Return a PathOs with resolved_path as its value.

  6. Return a PathOs with abs_path as its value.

---

function PathOsToCfs(os_path: PathOs) -> PathCfs
  errors:
    - ResolvesOutsideRoot
    - CannotDetermineRoot (propagated from PathGetProjectRoot)

  1. Call PathGetProjectRoot.
     If it raises an error, propagate that error immediately without continuing.
     Store the result as root.

  2. If os_path.value refers to an existing filesystem entry:
       Resolve any symlinks in os_path.value to obtain resolved_path.
     Else:
       Set resolved_path to os_path.value.

  3. If resolved_path does not start with root.value, raise error "resolves outside root".

  4. Compute the relative portion of resolved_path by removing the root.value prefix
     and any leading path separator.
     Store the result as rel_path.

  5. Replace all OS-native path separator characters in rel_path with "/".

  6. Return a PathCfs with the result as its value.
