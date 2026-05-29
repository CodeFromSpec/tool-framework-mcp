<!-- code-from-spec: ROOT/functional/tests/os/path_utils@Pn-4eDp7MlCiievotb8mG1NLLwA -->

# Test Specification: path_utils

---

## PathValidateCfs

---

### TC-PV-01: Valid simple relative path

**Setup:** none

**Action:** Call `PathValidateCfs` with `"internal/config/config.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-02: Valid nested path

**Setup:** none

**Action:** Call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-03: Valid single filename

**Setup:** none

**Action:** Call `PathValidateCfs` with `"main.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-04: Accepts path with dot segment

**Setup:** none

**Action:** Call `PathValidateCfs` with `"internal/./config/config.go"`.

**Expected outcome:** No error is returned. A single dot segment resolves
harmlessly and does not constitute a traversal.

---

### TC-PV-05: Accepts traversal that resolves within root

**Setup:** none

**Action:** Call `PathValidateCfs` with `"a/b/../c"`.

**Expected outcome:** No error is returned. After normalization the path
becomes `"a/c"`, which contains no `..` components.

---

### TC-PV-06: Accepts path with trailing slash

**Setup:** none

**Action:** Call `PathValidateCfs` with `"internal/config/"`.

**Expected outcome:** No error is returned.

---

### TC-PV-07: Accepts path with duplicate slashes

**Setup:** none

**Action:** Call `PathValidateCfs` with `"internal//config//file.go"`.

**Expected outcome:** No error is returned.

---

### TC-PV-08: Rejects empty string

**Setup:** none

**Action:** Call `PathValidateCfs` with `""`.

**Expected outcome:** An error matching `"path is empty"` is returned.

---

### TC-PV-09: Rejects absolute path with leading slash

**Setup:** none

**Action:** Call `PathValidateCfs` with `"/etc/passwd"`.

**Expected outcome:** An error matching `"path is absolute"` is returned.

---

### TC-PV-10: Rejects absolute path with drive letter

**Setup:** none

**Action:** Call `PathValidateCfs` with `"C:/Windows/system32"`.

**Expected outcome:** An error matching `"path is absolute"` is returned.

---

### TC-PV-11: Rejects backslash

**Setup:** none

**Action:** Call `PathValidateCfs` with `"internal\config\config.go"`.

**Expected outcome:** An error matching `"path contains backslash"` is returned.

---

### TC-PV-12: Rejects simple traversal

**Setup:** none

**Action:** Call `PathValidateCfs` with `"../../etc/passwd"`.

**Expected outcome:** An error matching `"directory traversal"` is returned.

---

### TC-PV-13: Rejects embedded traversal

**Setup:** none

**Action:** Call `PathValidateCfs` with `"internal/../../outside/file.go"`.

**Expected outcome:** An error matching `"directory traversal"` is returned.

---

## PathCfsToOs

---

### TC-CO-01: Converts valid path that exists

**Setup:** A file exists at `"internal/config/config.go"` inside the
project root.

**Action:** Call `PathCfsToOs` with a `PathCfs` whose value is
`"internal/config/config.go"`.

**Expected outcome:** No error is returned. The resulting `PathOs` value is
absolute and ends with the OS-specific equivalent of
`"internal/config/config.go"` (i.e. with the correct separator for the
current OS).

---

### TC-CO-02: Converts valid path that does not exist

**Setup:** No file exists at `"internal/newdir/newfile.go"` inside the
project root.

**Action:** Call `PathCfsToOs` with a `PathCfs` whose value is
`"internal/newdir/newfile.go"`.

**Expected outcome:** No error is returned. The resulting `PathOs` value is
absolute and ends with the OS-specific equivalent of
`"internal/newdir/newfile.go"`.

---

### TC-CO-03: Converts path with duplicate slashes

**Setup:** none

**Action:** Call `PathCfsToOs` with a `PathCfs` whose value is
`"internal//config.go"`.

**Expected outcome:** No error is returned. The path is normalized — the
resulting `PathOs` does not contain duplicate separators.

---

### TC-CO-04: Rejects invalid CfsPath

**Setup:** none

**Action:** Call `PathCfsToOs` with a `PathCfs` whose value is
`"../../etc/passwd"`.

**Expected outcome:** An error matching `"directory traversal"` is returned.
No `PathOs` is produced.

---

### TC-CO-05: Rejects symlink escaping project root

**Setup:** Create a symlink inside the project root whose target is a
directory outside the project root.

**Action:** Call `PathCfsToOs` with a `PathCfs` that resolves through
that symlink (e.g. `"<symlink-name>/sensitive-file"`).

**Expected outcome:** An error matching `"resolves outside root"` is returned.

---

### TC-CO-06: Roundtrip — CfsToOs then OsToCfs

**Setup:** none

**Action:**
1. Call `PathCfsToOs` with `"internal/config/config.go"` to obtain a
   `PathOs`.
2. Call `PathOsToCfs` with that `PathOs`.

**Expected outcome:** No errors at either step. The final `PathCfs` value
equals `"internal/config/config.go"`.

---

## PathOsToCfs

---

### TC-OC-01: Converts valid OS path that exists

**Setup:** A file exists inside the project root. Its absolute OS path is
known.

**Action:** Call `PathOsToCfs` with a `PathOs` whose value is the
absolute OS path to that file.

**Expected outcome:** No error is returned. The resulting `PathCfs` value
uses forward slashes, is relative (no leading slash, no drive letter), and
correctly represents the path relative to the project root.

---

### TC-OC-02: Converts valid OS path that does not exist

**Setup:** Construct an absolute OS path to a location inside the project
root that does not correspond to an existing file.

**Action:** Call `PathOsToCfs` with a `PathOs` whose value is that
constructed path.

**Expected outcome:** No error is returned. The resulting `PathCfs` value
uses forward slashes and is correctly relative to the project root.

---

### TC-OC-03: Result uses forward slashes

**Setup:** none

**Action:** Call `PathOsToCfs` with a valid absolute OS path that is within
the project root.

**Expected outcome:** No error is returned. The resulting `PathCfs` value
contains no backslash characters, regardless of the OS.

---

### TC-OC-04: Symlink within root resolving within root

**Setup:** Create a symlink inside the project root whose target is another
location also inside the project root.

**Action:** Call `PathOsToCfs` with the absolute OS path of that symlink.

**Expected outcome:** No error is returned. A valid `PathCfs` is returned.

---

### TC-OC-05: Rejects path outside project root

**Setup:** none

**Action:** Call `PathOsToCfs` with an absolute OS path that is outside the
project root (e.g. `"/tmp/outside/file.txt"` on Unix or an equivalent on
Windows).

**Expected outcome:** An error matching `"resolves outside root"` is returned.

---

## PathGetProjectRoot

---

### TC-GR-01: Returns an absolute path

**Setup:** none

**Action:** Call `PathGetProjectRoot`.

**Expected outcome:** No error is returned. The resulting `PathOs` value is
non-empty and is an absolute path (starts with `/` on Unix, or a drive
letter followed by `\` on Windows).

---

### TC-GR-02: Matches working directory

**Setup:** none

**Action:** Call `PathGetProjectRoot`. Also obtain the current working
directory of the process by any available means.

**Expected outcome:** No error is returned. The `PathOs` value returned by
`PathGetProjectRoot` corresponds to the current working directory of the
process.
