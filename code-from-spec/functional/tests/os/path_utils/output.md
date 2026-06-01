<!-- code-from-spec: ROOT/functional/tests/os/path_utils@2hMhcWS4DZu0j2KMtbLoyCCta-g -->

# Test Specification: PathUtils

---

## PathValidateCfs

---

### TC-PV-01: Valid simple relative path

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"internal/config/config.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-02: Valid nested path

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-03: Valid single filename

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"main.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-04: Accepts path with dot segment

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"internal/./config/config.go"`.

**Expected outcome:** No error is returned. A single dot segment resolves
harmlessly and does not constitute a traversal.

---

### TC-PV-05: Accepts traversal that resolves within root

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"a/b/../c"`.

**Expected outcome:** No error is returned. After normalization the path
becomes `"a/c"`, which contains no `..` components.

---

### TC-PV-06: Accepts path with trailing slash

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"internal/config/"`.

**Expected outcome:** No error is returned.

---

### TC-PV-07: Accepts path with duplicate slashes

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"internal//config//file.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-08: Rejects empty string

**Setup:** None.

**Action:** Call `PathValidateCfs` with `""`.

**Expected outcome:** Error `PathEmpty` is returned.

---

### TC-PV-09: Rejects absolute path with leading slash

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"/etc/passwd"`.

**Expected outcome:** Error `PathAbsolute` is returned.

---

### TC-PV-10: Rejects absolute path with drive letter

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"C:/Windows/system32"`.

**Expected outcome:** Error `PathAbsolute` is returned.

---

### TC-PV-11: Rejects backslash

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"internal\config\config.go"`.

**Expected outcome:** Error `PathContainsBackslash` is returned.

---

### TC-PV-12: Rejects simple traversal

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"../../etc/passwd"`.

**Expected outcome:** Error `DirectoryTraversal` is returned.

---

### TC-PV-13: Rejects embedded traversal

**Setup:** None.

**Action:** Call `PathValidateCfs` with `"internal/../../outside/file.go"`.

**Expected outcome:** Error `DirectoryTraversal` is returned.

---

## PathCfsToOs

---

### TC-CO-01: Converts valid path that exists

**Setup:** Ensure a file exists at `"internal/config/config.go"` inside
the project root.

**Action:** Call `PathCfsToOs` with `"internal/config/config.go"`.

**Expected outcome:** No error is returned. The resulting `PathOs` value
is absolute and ends with the OS-specific equivalent of
`internal/config/config.go` (using the OS path separator).

---

### TC-CO-02: Converts valid path that does not exist

**Setup:** Ensure no file exists at `"internal/newdir/newfile.go"` inside
the project root.

**Action:** Call `PathCfsToOs` with `"internal/newdir/newfile.go"`.

**Expected outcome:** No error is returned. The resulting `PathOs` value
is absolute and ends with the OS-specific equivalent of
`internal/newdir/newfile.go`. The target file does not need to exist for
the conversion to succeed.

---

### TC-CO-03: Converts path with duplicate slashes

**Setup:** None.

**Action:** Call `PathCfsToOs` with `"internal//config.go"`.

**Expected outcome:** No error is returned. The path is normalized; the
resulting `PathOs` does not contain duplicate separators.

---

### TC-CO-04: Rejects invalid CfsPath

**Setup:** None.

**Action:** Call `PathCfsToOs` with `"../../etc/passwd"`.

**Expected outcome:** Error `DirectoryTraversal` is returned. No `PathOs`
is produced.

---

### TC-CO-05: Rejects symlink escaping project root

**Setup:** Create a symlink inside the project root whose target is a
directory outside the project root.

**Action:** Call `PathCfsToOs` with a CFS path that traverses through
that symlink.

**Expected outcome:** Error `ResolvesOutsideRoot` is returned.

---

### TC-CO-06: Roundtrip — CfsToOs then OsToCfs

**Setup:** None.

**Action:**
1. Call `PathCfsToOs` with `"internal/config/config.go"` to obtain a
   `PathOs`.
2. Call `PathOsToCfs` with that `PathOs`.

**Expected outcome:** No error is returned at either step. The final
`PathCfs` value equals `"internal/config/config.go"`.

---

## PathOsToCfs

---

### TC-OC-01: Converts valid OS path that exists

**Setup:** Create a file inside the project root. Record its absolute OS
path.

**Action:** Call `PathOsToCfs` with that absolute OS path.

**Expected outcome:** No error is returned. The resulting `PathCfs` uses
forward slashes and is relative to the project root.

---

### TC-OC-02: Converts valid OS path that does not exist

**Setup:** Construct an absolute OS path to a location inside the project
root that does not correspond to an existing file or directory.

**Action:** Call `PathOsToCfs` with that absolute OS path.

**Expected outcome:** No error is returned. The resulting `PathCfs` uses
forward slashes and is relative to the project root. The target does not
need to exist for the conversion to succeed.

---

### TC-OC-03: Result uses forward slashes

**Setup:** Construct a valid absolute OS path to a location inside the
project root.

**Action:** Call `PathOsToCfs` with that path on any OS.

**Expected outcome:** No error is returned. The resulting `PathCfs` value
contains no backslash characters.

---

### TC-OC-04: Symlink within root resolving within root

**Setup:** Create a symlink inside the project root whose target is also
inside the project root.

**Action:** Call `PathOsToCfs` with the absolute OS path of the symlink.

**Expected outcome:** No error is returned. A valid `PathCfs` is returned.

---

### TC-OC-05: Rejects path outside project root

**Setup:** Obtain an absolute OS path that is outside the project root
(e.g. a path to a system directory or a temp directory outside the
working directory tree).

**Action:** Call `PathOsToCfs` with that path.

**Expected outcome:** Error `ResolvesOutsideRoot` is returned.

---

## PathGetProjectRoot

---

### TC-PR-01: Returns an absolute path

**Setup:** None.

**Action:** Call `PathGetProjectRoot`.

**Expected outcome:** No error is returned. The resulting `PathOs` value
is non-empty and is an absolute path (starts with `/` on Unix, or a drive
letter followed by `\` on Windows).

---

### TC-PR-02: Matches working directory

**Setup:** Record the current working directory of the process before
calling the function.

**Action:** Call `PathGetProjectRoot`.

**Expected outcome:** No error is returned. The resulting `PathOs`
corresponds to the current working directory recorded in setup.
