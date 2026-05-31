<!-- code-from-spec: ROOT/functional/logic/os/path_utils@AbxijKGAOclUZ47fL-U4A3sAvtc -->

# PathUtils — Pseudocode

## Records

```
record PathCfs
  value: string

record PathOs
  value: string
```

---

## function PathGetProjectRoot() -> PathOs

  1. Read the working directory of the current process.
     If the working directory cannot be read, raise error "cannot determine project root".

  2. Return a PathOs whose value is the working directory path.

  errors:
    - CannotDetermineRoot: the working directory cannot be read.

---

## function PathValidateCfs(value: string)

  1. If value is empty, raise error "path is empty".

  2. If value starts with "/" or starts with a drive letter pattern
     (a single letter followed by ":"), raise error "path is absolute".

  3. If value contains any "\" character, raise error "path contains backslash".

  4. Normalize the path by resolving "." and ".." components.

  5. For each component in the normalized path:
       If the component is "..", raise error "directory traversal".

  errors:
    - PathEmpty: the path value is empty.
    - PathAbsolute: the path starts with "/" or a drive letter like "C:".
    - PathContainsBackslash: the path contains "\" characters.
    - DirectoryTraversal: the path contains ".." components after normalization.

---

## function PathCfsToOs(cfs_path: PathCfs) -> PathOs

  1. Call PathValidateCfs with cfs_path.value.
     If it raises an error, propagate that error to the caller.

  2. Call PathGetProjectRoot to obtain the project root as a PathOs.
     If it raises an error, propagate that error to the caller.

  3. Replace all "/" characters in cfs_path.value with the OS path separator.

  4. Join the project root value and the converted path to form an absolute path.

  5. If the path exists on disk:
       Resolve symlinks on the absolute path.
       If the resolved path does not start with the project root value,
         raise error "resolves outside root".

  6. Return a PathOs whose value is the absolute path from step 4.

  errors:
    - ResolvesOutsideRoot: after resolving symlinks, the path is outside the project root.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (PathUtils.*): propagated from PathGetProjectRoot.

---

## function PathOsToCfs(os_path: PathOs) -> PathCfs

  1. Call PathGetProjectRoot to obtain the project root as a PathOs.
     If it raises an error, propagate that error to the caller.

  2. If os_path.value exists on disk:
       Resolve symlinks on os_path.value.
       Use the resolved path as the working value for subsequent steps.
     Else:
       Use os_path.value as the working value.

  3. If the working value does not start with the project root value,
     raise error "resolves outside root".

  4. Compute the relative path by removing the project root prefix
     (and any leading separator) from the working value.

  5. Replace all OS path separator characters in the relative path
     with "/".

  6. Return a PathCfs whose value is the result from step 5.

  errors:
    - ResolvesOutsideRoot: the path is not within the project root.
    - (PathUtils.*): propagated from PathGetProjectRoot.
```
