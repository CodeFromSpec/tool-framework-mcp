<!-- code-from-spec: ROOT/functional/logic/os/path_utils@j6ywRj5HVrT0-qP-mnfyrigoqfo -->

# path_utils

namespace: pathutils

---

## Records

record PathCfs
  value: string

record PathOs
  value: string

---

## Functions

### PathGetProjectRoot() -> PathOs

Returns the current working directory of the process as a PathOs (absolute,
OS-native format).

  1. Read the working directory of the current process.
     If the working directory cannot be read, raise error "cannot determine
     project root".

  2. Return a PathOs with that absolute path as its value.

Errors:
  - CannotDetermineRoot: the working directory cannot be read.

---

### PathValidateCfs(value: string)

Validates that a string conforms to the PathCfs format rules. Raises an error
describing the first violation found. Does not check whether the file exists.

  1. If value is empty, raise error "path is empty".

  2. If value starts with "/" or matches a drive-letter pattern (e.g. "C:"),
     raise error "path is absolute".

  3. If value contains "\", raise error "path contains backslash".

  4. Normalize the path by resolving "." and ".." components.

  5. For each component of the normalized path:
       If the component equals "..", raise error "directory traversal detected".

Errors:
  - PathEmpty: the path value is empty.
  - PathAbsolute: the path starts with "/" or a drive letter like "C:".
  - PathContainsBackslash: the path contains "\" characters.
  - DirectoryTraversal: the path contains ".." components after normalization.

---

### PathCfsToOs(cfs_path: PathCfs) -> PathOs

Validates a PathCfs and converts it to an absolute PathOs. This is the single
entry point for going from framework paths to OS paths. Never sanitizes —
rejects invalid paths. Never creates or modifies files.

  1. Call PathValidateCfs with cfs_path.value.
     If it raises any error, propagate that error to the caller.

  2. Obtain the project root by calling PathGetProjectRoot.
     If it raises any error, propagate that error to the caller.

  3. Replace all forward slash "/" characters in cfs_path.value with the OS
     path separator.

  4. Join the project root path with the converted relative path to form a
     single absolute path.

  5. Check whether the resulting path exists on the filesystem.
     If it exists:
       a. Resolve all symlinks in the path to obtain the real absolute path.
       b. Verify that the resolved path starts with the project root path
          (using the project root as a prefix, including its trailing
          separator, to avoid partial directory name matches).
          If it does not, raise error "resolves outside root".

  6. Return a PathOs whose value is the absolute joined path from step 4
     (pre-symlink-resolution), confirming containment has passed.

Errors:
  - ResolvesOutsideRoot: after resolving symlinks, the path is outside the
    project root.
  - (propagated from PathValidateCfs): PathEmpty, PathAbsolute,
    PathContainsBackslash, DirectoryTraversal.
  - (propagated from PathGetProjectRoot): CannotDetermineRoot.

---

### PathOsToCfs(os_path: PathOs) -> PathCfs

Converts an absolute PathOs to a PathCfs relative to the project root. Used
by components that receive paths from the OS (e.g. directory listing). Never
creates or modifies files.

  1. Obtain the project root by calling PathGetProjectRoot.
     If it raises any error, propagate that error to the caller.

  2. Check whether os_path.value exists on the filesystem.
     If it exists:
       a. Resolve all symlinks in os_path.value to obtain the real absolute
          path.
       b. Use the resolved path as the working path for the remaining steps.
     If it does not exist, use os_path.value as the working path.

  3. Verify that the working path starts with the project root path (using
     the project root as a prefix, including its trailing separator, to avoid
     partial directory name matches).
     If it does not, raise error "resolves outside root".

  4. Compute the relative portion of the working path by removing the project
     root prefix.

  5. Replace all OS path separator characters in the relative path with
     forward slashes "/".

  6. Return a PathCfs with the resulting relative path as its value.

Errors:
  - ResolvesOutsideRoot: the path is not within the project root.
  - (propagated from PathGetProjectRoot): CannotDetermineRoot.
