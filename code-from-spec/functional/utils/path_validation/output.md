<!-- code-from-spec: ROOT/functional/utils/path_validation@PENDING -->

function ValidatePath(relative_path, project_root) -> void

  1. If relative_path is empty,
     raise error "path is empty".

  2. If relative_path starts with "/" or matches a drive letter
     pattern (a letter followed by ":"),
     raise error "path is absolute".

  3. Normalize the path:
     a. Replace all backslash separators with forward slashes.
     b. Resolve any "." components (remove them).
     c. Resolve any ".." components by removing the preceding
        component.

  4. After normalization, check each component of the path.
     If any component is "..",
     raise error "directory traversal".

  5. Join project_root and the normalized relative_path to
     form the full absolute path.

  6. Resolve any symbolic links in the full absolute path
     to obtain the real path.

  7. Resolve any symbolic links in project_root to obtain
     the real project root.

  8. If the real path does not start with the real project root,
     raise error "resolves outside root".

  9. Return success.
