<!-- code-from-spec: SPEC/functional/tests/os/path_utils@jCHV9AKMXZUZj7oxgX_Z9uj80uM -->

## Test suite: PathUtils

---

### PathValidateCfs

---

#### TC-PV-01: Valid simple relative path

Actions:
1. Call `PathValidateCfs` with `"internal/config/config.go"`.

Expected outcome:
- No error is raised.

---

#### TC-PV-02: Valid nested path

Actions:
1. Call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.

Expected outcome:
- No error is raised.

---

#### TC-PV-03: Valid single filename

Actions:
1. Call `PathValidateCfs` with `"main.go"`.

Expected outcome:
- No error is raised.

---

#### TC-PV-04: Accepts path with dot segment

Actions:
1. Call `PathValidateCfs` with `"internal/./config/config.go"`.

Expected outcome:
- No error is raised.
  The dot component resolves harmlessly and does not violate any rule.

---

#### TC-PV-05: Accepts traversal that resolves within root

Actions:
1. Call `PathValidateCfs` with `"a/b/../c"`.

Expected outcome:
- No error is raised.
  After normalization this becomes `"a/c"`, which contains no `..` components.

---

#### TC-PV-06: Accepts path with trailing slash

Actions:
1. Call `PathValidateCfs` with `"internal/config/"`.

Expected outcome:
- No error is raised.

---

#### TC-PV-07: Accepts path with duplicate slashes

Actions:
1. Call `PathValidateCfs` with `"internal//config//file.go"`.

Expected outcome:
- No error is raised.

---

#### TC-PV-08: Rejects empty string

Actions:
1. Call `PathValidateCfs` with `""`.

Expected outcome:
- Error PathEmpty is raised.

---

#### TC-PV-09: Rejects absolute path with leading slash

Actions:
1. Call `PathValidateCfs` with `"/etc/passwd"`.

Expected outcome:
- Error PathAbsolute is raised.

---

#### TC-PV-10: Rejects absolute path with drive letter

Actions:
1. Call `PathValidateCfs` with `"C:/Windows/system32"`.

Expected outcome:
- Error PathAbsolute is raised.

---

#### TC-PV-11: Rejects backslash

Actions:
1. Call `PathValidateCfs` with `"internal\config\config.go"`.

Expected outcome:
- Error PathContainsBackslash is raised.

---

#### TC-PV-12: Rejects simple traversal

Actions:
1. Call `PathValidateCfs` with `"../../etc/passwd"`.

Expected outcome:
- Error DirectoryTraversal is raised.

---

#### TC-PV-13: Rejects embedded traversal

Actions:
1. Call `PathValidateCfs` with `"internal/../../outside/file.go"`.

Expected outcome:
- Error DirectoryTraversal is raised.

---

### PathCfsToOs

---

#### TC-CO-01: Converts valid path that exists

Setup:
- A file exists at `"internal/config/config.go"` inside the project root.

Actions:
1. Call `PathCfsToOs` with a `PathCfs` whose value is `"internal/config/config.go"`.

Expected outcome:
- No error is raised.
- The result is a `PathOs` whose value is absolute.
- The result path ends with the OS-specific equivalent of `"internal/config/config.go"`.

---

#### TC-CO-02: Converts valid path that does not exist

Setup:
- No file exists at `"internal/newdir/newfile.go"` inside the project root.

Actions:
1. Call `PathCfsToOs` with a `PathCfs` whose value is `"internal/newdir/newfile.go"`.

Expected outcome:
- No error is raised.
- The result is a `PathOs` whose value is absolute.
- The result path ends with the OS-specific equivalent of `"internal/newdir/newfile.go"`.

---

#### TC-CO-03: Converts path with duplicate slashes

Actions:
1. Call `PathCfsToOs` with a `PathCfs` whose value is `"internal//config.go"`.

Expected outcome:
- No error is raised.
- The result is a normalized absolute `PathOs`.

---

#### TC-CO-04: Rejects invalid CfsPath — directory traversal

Actions:
1. Call `PathCfsToOs` with a `PathCfs` whose value is `"../../etc/passwd"`.

Expected outcome:
- Error DirectoryTraversal is raised.
- No `PathOs` is returned.

---

#### TC-CO-05: Rejects symlink escaping project root

Setup:
- A directory exists outside the project root.
- A file exists inside that outside directory (so the symlink target is a real file on disk).
- A symlink exists inside the project root pointing to that outside file.

Actions:
1. Call `PathCfsToOs` with a `PathCfs` whose value resolves through the symlink.

Expected outcome:
- Error ResolvesOutsideRoot is raised.

---

#### TC-CO-06: Roundtrip — CfsToOs then OsToCfs

Actions:
1. Call `PathCfsToOs` with a `PathCfs` whose value is `"internal/config/config.go"`.
   Capture the resulting `PathOs`.
2. Call `PathOsToCfs` with that `PathOs`.
   Capture the resulting `PathCfs`.

Expected outcome:
- No error is raised at either step.
- The final `PathCfs` value equals `"internal/config/config.go"`.

---

### PathOsToCfs

---

#### TC-OC-01: Converts valid OS path that exists

Setup:
- A file exists inside the project root.
  Let its absolute OS path be <absolute_path_inside_root>.

Actions:
1. Call `PathOsToCfs` with a `PathOs` whose value is <absolute_path_inside_root>.

Expected outcome:
- No error is raised.
- The result is a `PathCfs` with forward slashes.
- The result value is relative to the project root.

---

#### TC-OC-02: Converts valid OS path that does not exist

Setup:
- Construct an absolute OS path to a file that does not exist but is within the project root.
  Let this path be <nonexistent_absolute_path>.

Actions:
1. Call `PathOsToCfs` with a `PathOs` whose value is <nonexistent_absolute_path>.

Expected outcome:
- No error is raised.
- The result is a `PathCfs` with forward slashes.
- The result value is relative to the project root.

---

#### TC-OC-03: Result uses forward slashes

Actions:
1. Call `PathOsToCfs` with a valid absolute OS path inside the project root.

Expected outcome:
- No error is raised.
- The resulting `PathCfs` value contains no backslash characters.

---

#### TC-OC-04: Symlink within root resolving within root

Setup:
- A file exists inside the project root (the symlink target).
- A symlink exists inside the project root pointing to that file.

Actions:
1. Call `PathOsToCfs` with a `PathOs` whose value is the absolute OS path of the symlink.

Expected outcome:
- No error is raised.
- The result is a valid `PathCfs` relative to the project root.

---

#### TC-OC-05: Rejects path outside project root

Actions:
1. Call `PathOsToCfs` with a `PathOs` whose value is an absolute OS path outside the project root.

Expected outcome:
- Error ResolvesOutsideRoot is raised.

---

### PathGetProjectRoot

---

#### TC-PGR-01: Returns an absolute path

Actions:
1. Call `PathGetProjectRoot`.

Expected outcome:
- No error is raised.
- The result is a `PathOs` whose value is a non-empty absolute path.

---

#### TC-PGR-02: Matches working directory

Actions:
1. Call `PathGetProjectRoot`.

Expected outcome:
- No error is raised.
- The result corresponds to the current working directory of the process.
