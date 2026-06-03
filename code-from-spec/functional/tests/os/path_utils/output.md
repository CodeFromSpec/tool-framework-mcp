<!-- code-from-spec: ROOT/functional/tests/os/path_utils@AGeNGnfMwkR78gBlmZ1kic4RMhM -->

## PathValidateCfs

### Valid simple relative path

Actions: Call `PathValidateCfs` with `"internal/config/config.go"`.
Expected: No error.

### Valid nested path

Actions: Call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.
Expected: No error.

### Valid single filename

Actions: Call `PathValidateCfs` with `"main.go"`.
Expected: No error.

### Accepts path with dot segment

Actions: Call `PathValidateCfs` with `"internal/./config/config.go"`.
Expected: No error — dot resolves harmlessly.

### Accepts traversal that resolves within root

Actions: Call `PathValidateCfs` with `"a/b/../c"`.
Expected: No error — after normalization this becomes `"a/c"` which has no `..` components.

### Accepts path with trailing slash

Actions: Call `PathValidateCfs` with `"internal/config/"`.
Expected: No error.

### Accepts path with duplicate slashes

Actions: Call `PathValidateCfs` with `"internal//config//file.go"`.
Expected: No error.

### Rejects empty string

Actions: Call `PathValidateCfs` with `""`.
Expected: Error PathEmpty.

### Rejects absolute path with leading slash

Actions: Call `PathValidateCfs` with `"/etc/passwd"`.
Expected: Error PathAbsolute.

### Rejects absolute path with drive letter

Actions: Call `PathValidateCfs` with `"C:/Windows/system32"`.
Expected: Error PathAbsolute.

### Rejects backslash

Actions: Call `PathValidateCfs` with `"internal\config\config.go"`.
Expected: Error PathContainsBackslash.

### Rejects simple traversal

Actions: Call `PathValidateCfs` with `"../../etc/passwd"`.
Expected: Error DirectoryTraversal.

### Rejects embedded traversal

Actions: Call `PathValidateCfs` with `"internal/../../outside/file.go"`.
Expected: Error DirectoryTraversal.

---

## PathCfsToOs

### Converts valid path that exists

Setup: Create a file at `"internal/config/config.go"` inside the project root.
Actions: Call `PathCfsToOs` with `"internal/config/config.go"`.
Expected: Success — a `PathOs` that is absolute and ends with the OS-specific equivalent of that path.

### Converts valid path that does not exist

Setup: No file at `"internal/newdir/newfile.go"`.
Actions: Call `PathCfsToOs` with `"internal/newdir/newfile.go"`.
Expected: Success — a `PathOs` that is absolute and ends with the OS-specific equivalent of that path.

### Converts path with duplicate slashes

Actions: Call `PathCfsToOs` with `"internal//config.go"`.
Expected: Success — the path is normalized.

### Rejects invalid CfsPath

Actions: Call `PathCfsToOs` with `"../../etc/passwd"`.
Expected: Error DirectoryTraversal.

### Rejects symlink escaping project root

Setup: Create a directory outside the project root with a file inside it. Create a symlink inside the project root pointing to that outside file.
Actions: Call `PathCfsToOs` with a path through the symlink.
Expected: Error ResolvesOutsideRoot.

### Roundtrip: CfsToOs then OsToCfs

Actions:
  1. Call `PathCfsToOs` with `"internal/config/config.go"` to get a `PathOs`.
  2. Call `PathOsToCfs` with that `PathOs`.
Expected: The final `PathCfs` value equals `"internal/config/config.go"`.

---

## PathOsToCfs

### Converts valid OS path that exists

Setup: Create a file inside the project root. Construct its absolute OS path.
Actions: Call `PathOsToCfs` with that absolute OS path.
Expected: Success — a `PathCfs` with forward slashes, relative to the project root.

### Converts valid OS path that does not exist

Setup: Construct an absolute OS path to a file that does not exist but is within the project root.
Actions: Call `PathOsToCfs` with that path.
Expected: Success — a `PathCfs` with forward slashes relative to the project root.

### Result uses forward slashes

Actions: Call `PathOsToCfs` with a valid absolute OS path within the project root.
Expected: The resulting `PathCfs` contains no backslashes.

### Symlink within root resolving within root

Setup: Create a file inside the project root. Create a symlink inside the project root pointing to that file.
Actions: Call `PathOsToCfs` with the symlink path.
Expected: Success.

### Rejects path outside project root

Actions: Call `PathOsToCfs` with an absolute OS path that is outside the project root.
Expected: Error ResolvesOutsideRoot.

---

## PathGetProjectRoot

### Returns an absolute path

Actions: Call `PathGetProjectRoot`.
Expected: The result is a `PathOs` that is a non-empty absolute path.

### Matches working directory

Actions: Call `PathGetProjectRoot`.
Expected: The result corresponds to the current working directory of the process.
