<!-- code-from-spec: ROOT/functional/tests/os/path_utils@Kns95t-G9NK7KOjFEaZFAYQt5vk -->

## PathValidateCfs

### valid simple relative path

Setup: none.

Action: call `PathValidateCfs` with `"internal/config/config.go"`.

Expected outcome: no error is raised.

---

### valid nested path

Setup: none.

Action: call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.

Expected outcome: no error is raised.

---

### valid single filename

Setup: none.

Action: call `PathValidateCfs` with `"main.go"`.

Expected outcome: no error is raised.

---

### accepts path with dot segment

Setup: none.

Action: call `PathValidateCfs` with `"internal/./config/config.go"`.

Expected outcome: no error is raised. A dot segment resolves harmlessly
and does not escape the root.

---

### accepts traversal that resolves within root

Setup: none.

Action: call `PathValidateCfs` with `"a/b/../c"`.

Expected outcome: no error is raised. After normalization the path becomes
`"a/c"`, which contains no `..` components.

---

### accepts path with trailing slash

Setup: none.

Action: call `PathValidateCfs` with `"internal/config/"`.

Expected outcome: no error is raised.

---

### accepts path with duplicate slashes

Setup: none.

Action: call `PathValidateCfs` with `"internal//config//file.go"`.

Expected outcome: no error is raised.

---

### rejects empty string

Setup: none.

Action: call `PathValidateCfs` with `""`.

Expected outcome: error PathEmpty is raised.

---

### rejects absolute path with leading slash

Setup: none.

Action: call `PathValidateCfs` with `"/etc/passwd"`.

Expected outcome: error PathAbsolute is raised.

---

### rejects absolute path with drive letter

Setup: none.

Action: call `PathValidateCfs` with `"C:/Windows/system32"`.

Expected outcome: error PathAbsolute is raised.

---

### rejects backslash

Setup: none.

Action: call `PathValidateCfs` with `"internal\config\config.go"`.

Expected outcome: error PathContainsBackslash is raised.

---

### rejects simple traversal

Setup: none.

Action: call `PathValidateCfs` with `"../../etc/passwd"`.

Expected outcome: error DirectoryTraversal is raised.

---

### rejects embedded traversal

Setup: none.

Action: call `PathValidateCfs` with `"internal/../../outside/file.go"`.

Expected outcome: error DirectoryTraversal is raised.

---

## PathCfsToOs

### converts valid path that exists

Setup: create a file at `"internal/config/config.go"` inside the project
root.

Action: call `PathCfsToOs` with `"internal/config/config.go"`.

Expected outcome: no error. The returned `PathOs` is absolute and its
path ends with the OS-specific equivalent of
`"internal/config/config.go"`.

---

### converts valid path that does not exist

Setup: ensure no file exists at `"internal/newdir/newfile.go"` inside the
project root.

Action: call `PathCfsToOs` with `"internal/newdir/newfile.go"`.

Expected outcome: no error. The returned `PathOs` is absolute and its
path ends with the OS-specific equivalent of
`"internal/newdir/newfile.go"`.

---

### converts path with duplicate slashes

Setup: none.

Action: call `PathCfsToOs` with `"internal//config.go"`.

Expected outcome: no error. The path is normalized and the returned
`PathOs` is absolute.

---

### rejects invalid CfsPath

Setup: none.

Action: call `PathCfsToOs` with `"../../etc/passwd"`.

Expected outcome: error DirectoryTraversal is raised. No `PathOs` is
returned.

---

### rejects symlink escaping project root

Setup: create a directory outside the project root and a file inside it.
Create a symlink inside the project root that points to that outside file.
Let `<link-name>` be the name of the symlink file within the project root.

Action: call `PathCfsToOs` with `"<link-name>"`.

Expected outcome: error ResolvesOutsideRoot is raised.

---

### roundtrip CfsToOs then OsToCfs

Setup: none.

Action:
1. Call `PathCfsToOs` with `"internal/config/config.go"` to get a
   `PathOs` value.
2. Call `PathOsToCfs` with that `PathOs` value.

Expected outcome: no errors at either step. The final `PathCfs` value
equals `"internal/config/config.go"`.

---

## PathOsToCfs

### converts valid OS path that exists

Setup: create a file inside the project root. Construct the absolute OS
path to that file.

Action: call `PathOsToCfs` with that absolute OS path.

Expected outcome: no error. The returned `PathCfs` uses forward slashes
and is relative to the project root.

---

### converts valid OS path that does not exist

Setup: construct an absolute OS path to a file that does not exist but
whose location is within the project root.

Action: call `PathOsToCfs` with that absolute OS path.

Expected outcome: no error. The returned `PathCfs` uses forward slashes
and is relative to the project root.

---

### result uses forward slashes

Setup: on any operating system, identify a valid absolute OS path within
the project root.

Action: call `PathOsToCfs` with that absolute OS path.

Expected outcome: no error. The `value` field of the returned `PathCfs`
contains no backslash characters.

---

### symlink within root resolving within root

Setup: create a file inside the project root. Create a symlink inside the
project root pointing to that file. Construct the absolute OS path to the
symlink.

Action: call `PathOsToCfs` with the absolute OS path of the symlink.

Expected outcome: no error. The returned `PathCfs` is relative to the
project root.

---

### rejects path outside project root

Setup: identify an absolute OS path that is outside the project root.

Action: call `PathOsToCfs` with that absolute OS path.

Expected outcome: error ResolvesOutsideRoot is raised.

---

## PathGetProjectRoot

### returns an absolute path

Setup: none.

Action: call `PathGetProjectRoot`.

Expected outcome: no error. The returned `PathOs` value is non-empty and
is an absolute path as recognized by the OS.

---

### matches working directory

Setup: note the current working directory of the process.

Action: call `PathGetProjectRoot`.

Expected outcome: no error. The returned `PathOs` value corresponds to
the current working directory.
