---
depends_on:
  - SPEC/golang/implementation/os/path_utils
output: internal/pathutils/pathutils_test.go
---

# SPEC/golang/tests/os/path_utils

# Agent

## Test cases

### PathValidateCfs

#### Valid simple relative path

Input: "internal/config/config.go". Expect: no error.

#### Valid nested path

Input: "cmd/framework-mcp/main.go". Expect: no error.

#### Valid single filename

Input: "main.go". Expect: no error.

#### Accepts path with dot segment

Input: "internal/./config/config.go". Expect: no error.

#### Accepts traversal that resolves within root

Input: "a/b/../c". Expect: no error (normalizes to
"a/c").

#### Accepts path with trailing slash

Input: "internal/config/". Expect: no error.

#### Accepts path with duplicate slashes

Input: "internal//config//file.go". Expect: no error.

#### Rejects empty string

Input: "". Expect: ErrPathEmpty.

#### Rejects absolute path with leading slash

Input: "/etc/passwd". Expect: ErrPathAbsolute.

#### Rejects absolute path with drive letter

Input: "C:/Windows/system32". Expect: ErrPathAbsolute.

#### Rejects backslash

Input: "internal\config\config.go".
Expect: ErrPathContainsBackslash.

#### Rejects simple traversal

Input: "../../etc/passwd".
Expect: ErrDirectoryTraversal.

#### Rejects embedded traversal

Input: "internal/../../outside/file.go".
Expect: ErrDirectoryTraversal.

### PathCfsToOs

#### Converts valid path that exists

Setup: create file at "internal/config/config.go".

Expected: PathOs is absolute, ends with OS-specific
equivalent.

#### Converts valid path that does not exist

Input: "internal/newdir/newfile.go" (no such file).

Expected: success, PathOs is absolute.

#### Converts path with duplicate slashes

Input: "internal//config.go". Expected: success,
normalized.

#### Rejects invalid CfsPath

Input: "../../etc/passwd".
Expected: ErrDirectoryTraversal.

#### Rejects symlink escaping project root

Setup: create file outside root, symlink inside root
pointing to it.

Expected: ErrResolvesOutsideRoot.

#### Roundtrip: CfsToOs then OsToCfs

Input: "internal/config/config.go".
CfsToOs → OsToCfs. Expected: equals original.

### PathOsToCfs

#### Converts valid OS path that exists

Setup: file inside project root.

Expected: PathCfs with forward slashes, relative to
root.

#### Converts valid OS path that does not exist

Setup: absolute OS path within root, no such file.

Expected: success, PathCfs with forward slashes.

#### Result uses forward slashes

Expected: no backslashes in result.

#### Symlink within root resolving within root

Setup: file inside root, symlink inside root pointing
to it.

Expected: success.

#### Rejects path outside project root

Input: absolute OS path outside root.

Expected: ErrResolvesOutsideRoot.

### PathGetProjectRoot

#### Returns an absolute path

Expected: non-empty absolute PathOs.

#### Matches working directory

Expected: corresponds to current working directory.

## Go-specific guidance

- The package name is `pathutils_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- For symlink tests, skip on platforms where symlinks
  are not supported.
