---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
output: internal/oslayer/oslayer_path_test.go
---

# SPEC/golang/tests/oslayer/path

# Agent

## Test cases

### ValidateCfsPath

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

### CfsPathToOs

#### Converts valid path that exists

Setup: create file at "internal/config/config.go".

Expected: OsPath is absolute, ends with OS-specific
equivalent.

#### Converts valid path that does not exist

Input: "internal/newdir/newfile.go" (no such file).

Expected: success, OsPath is absolute.

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

#### Roundtrip: CfsPathToOs then OsPathToCfs

Input: "internal/config/config.go".
CfsPathToOs → OsPathToCfs. Expected: equals original.

### OsPathToCfs

#### Converts valid OS path that exists

Setup: file inside project root.

Expected: CfsPath with forward slashes, relative to
root.

#### Converts valid OS path that does not exist

Setup: absolute OS path within root, no such file.

Expected: success, CfsPath with forward slashes.

#### Result uses forward slashes

Expected: no backslashes in result.

#### Symlink within root resolving within root

Setup: file inside root, symlink inside root pointing
to it.

Expected: success.

#### Rejects path outside project root

Input: absolute OS path outside root.

Expected: ErrResolvesOutsideRoot.

### GetProjectRoot

#### Returns an absolute path

Expected: non-empty absolute OsPath.

#### Matches working directory

Expected: corresponds to current working directory.

## Go-specific guidance

- The package name is `oslayer_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- For symlink tests, skip on platforms where symlinks
  are not supported.
