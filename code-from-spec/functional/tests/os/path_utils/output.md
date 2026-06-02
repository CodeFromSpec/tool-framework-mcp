<!-- code-from-spec: ROOT/functional/tests/os/path_utils@F2zflFQ2C4QHMhvjamR07L8ok4o -->

## PathValidateCfs

### valid simple relative path

Setup: none.

Action: call `PathValidateCfs` with `"internal/config/config.go"`.

Expected: no error is returned.

---

### valid nested path

Setup: none.

Action: call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.

Expected: no error is returned.

---

### valid single filename

Setup: none.

Action: call `PathValidateCfs` with `"main.go"`.

Expected: no error is returned.

---

### accepts path with dot segment

Setup: none.

Action: call `PathValidateCfs` with `"internal/./config/config.go"`.

Expected: no error is returned — a single dot resolves harmlessly.

---

### accepts traversal that resolves within root

Setup: none.

Action: call `PathValidateCfs` with `"a/b/../c"`.

Expected: no error is returned — after normalization the path becomes `"a/c"`, which contains no `..` components.

---

### accepts path with trailing slash

Setup: none.

Action: call `PathValidateCfs` with `"internal/config/"`.

Expected: no error is returned.

---

### accepts path with duplicate slashes

Setup: none.

Action: call `PathValidateCfs` with `"internal//config//file.go"`.

Expected: no error is returned.

---

### rejects empty string

Setup: none.

Action: call `PathValidateCfs` with `""`.

Expected: error PathEmpty is returned.

---

### rejects absolute path with leading slash

Setup: none.

Action: call `PathValidateCfs` with `"/etc/passwd"`.

Expected: error PathAbsolute is returned.

---

### rejects absolute path with drive letter

Setup: none.

Action: call `PathValidateCfs` with `"C:/Windows/system32"`.

Expected: error PathAbsolute is returned.

---

### rejects backslash

Setup: none.

Action: call `PathValidateCfs` with `"internal\config\config.go"`.

Expected: error PathContainsBackslash is returned.

---

### rejects simple traversal

Setup: none.

Action: call `PathValidateCfs` with `"../../etc/passwd"`.

Expected: error DirectoryTraversal is returned.

---

### rejects embedded traversal

Setup: none.

Action: call `PathValidateCfs` with `"internal/../../outside/file.go"`.

Expected: error DirectoryTraversal is returned.

---

## PathCfsToOs

### converts valid path that exists

Setup: a file exists at `"internal/config/config.go"` inside the project root.

Action: call `PathCfsToOs` with `"internal/config/config.go"`.

Expected: no error is returned. The result is a `PathOs` that is absolute and ends with the OS-specific equivalent of `internal/config/config.go`.

---

### converts valid path that does not exist

Setup: no file exists at `"internal/newdir/newfile.go"`.

Action: call `PathCfsToOs` with `"internal/newdir/newfile.go"`.

Expected: no error is returned. The result is a `PathOs` that is absolute and ends with the OS-specific equivalent of `internal/newdir/newfile.go`.

---

### converts path with duplicate slashes

Setup: none.

Action: call `PathCfsToOs` with `"internal//config.go"`.

Expected: no error is returned. The path is normalized before conversion.

---

### rejects invalid CfsPath

Setup: none.

Action: call `PathCfsToOs` with `"../../etc/passwd"`.

Expected: error DirectoryTraversal is returned. No conversion is attempted.

---

### rejects symlink escaping project root

Setup: a directory is created outside the project root with a file inside it (so the symlink target exists on disk). A symlink is created inside the project root pointing to that outside file.

Action: call `PathCfsToOs` with a path that traverses through that symlink.

Expected: error ResolvesOutsideRoot is returned.

---

### roundtrip: CfsToOs then OsToCfs

Setup: none.

Action:
1. Call `PathCfsToOs` with `"internal/config/config.go"` to obtain a `PathOs`.
2. Call `PathOsToCfs` with that `PathOs`.

Expected: no error at either step. The final `PathCfs` value equals `"internal/config/config.go"`.

---

## PathOsToCfs

### converts valid OS path that exists

Setup: a file is created inside the project root. Its absolute OS path is known.

Action: call `PathOsToCfs` with that absolute OS path.

Expected: no error is returned. The result is a `PathCfs` using forward slashes, relative to the project root, matching the file's location.

---

### converts valid OS path that does not exist

Setup: an absolute OS path is constructed for a file that does not exist but is within the project root.

Action: call `PathOsToCfs` with that path.

Expected: no error is returned. The result is a `PathCfs` using forward slashes, relative to the project root.

---

### result uses forward slashes

Setup: none.

Action: call `PathOsToCfs` with any valid absolute OS path within the project root.

Expected: no error is returned. The resulting `PathCfs` value contains no backslash characters, regardless of the operating system.

---

### symlink within root resolving within root

Setup: a file is created inside the project root (so the symlink target exists on disk). A symlink is created inside the project root pointing to that file.

Action: call `PathOsToCfs` with the absolute OS path of that symlink.

Expected: no error is returned. The result is a valid `PathCfs`.

---

### rejects path outside project root

Setup: none.

Action: call `PathOsToCfs` with an absolute OS path that is not within the project root.

Expected: error ResolvesOutsideRoot is returned.

---

## PathGetProjectRoot

### returns an absolute path

Setup: none.

Action: call `PathGetProjectRoot`.

Expected: no error is returned. The result is a `PathOs` that is non-empty and absolute.

---

### matches working directory

Setup: none.

Action: call `PathGetProjectRoot`.

Expected: no error is returned. The result corresponds to the current working directory of the process.
