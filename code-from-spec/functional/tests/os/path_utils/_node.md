---
depends_on:
  - ROOT/functional/logic/os/path_utils(interface)
outputs:
  - id: path_utils_tests
    path: code-from-spec/functional/tests/os/path_utils/output.md
---

# ROOT/functional/tests/os/path_utils

Test cases for path types and conversion functions.

# Public

## Test cases

### PathValidateCfs

#### Valid simple relative path

Call `PathValidateCfs` with `"internal/config/config.go"`.
Expect no error.

#### Valid nested path

Call `PathValidateCfs` with `"cmd/framework-mcp/main.go"`.
Expect no error.

#### Valid single filename

Call `PathValidateCfs` with `"main.go"`.
Expect no error.

#### Accepts path with dot segment

Call `PathValidateCfs` with `"internal/./config/config.go"`.
Expect no error (dot resolves harmlessly).

#### Accepts traversal that resolves within root

Call `PathValidateCfs` with `"a/b/../c"`. Expect no error —
after normalization this becomes `"a/c"` which has no
`..` components.

#### Accepts path with trailing slash

Call `PathValidateCfs` with `"internal/config/"`.
Expect no error.

#### Accepts path with duplicate slashes

Call `PathValidateCfs` with `"internal//config//file.go"`.
Expect no error.

#### Rejects empty string

Call `PathValidateCfs` with `""`.
Expect error PathEmpty.

#### Rejects absolute path with leading slash

Call `PathValidateCfs` with `"/etc/passwd"`.
Expect error PathAbsolute.

#### Rejects absolute path with drive letter

Call `PathValidateCfs` with `"C:/Windows/system32"`.
Expect error PathAbsolute.

#### Rejects backslash

Call `PathValidateCfs` with `"internal\config\config.go"`.
Expect error PathContainsBackslash.

#### Rejects simple traversal

Call `PathValidateCfs` with `"../../etc/passwd"`.
Expect error DirectoryTraversal.

#### Rejects embedded traversal

Call `PathValidateCfs` with `"internal/../../outside/file.go"`.
Expect error DirectoryTraversal.

### PathCfsToOs

#### Converts valid path that exists

Create a file at `"internal/config/config.go"` inside the
project root. Call `PathCfsToOs`. Expect a `PathOs` that
is absolute and ends with the OS-specific equivalent.

#### Converts valid path that does not exist

Call `PathCfsToOs` with `"internal/newdir/newfile.go"` —
no such file exists. Expect success — a `PathOs` that is
absolute and ends with the OS-specific equivalent.

#### Converts path with duplicate slashes

Call `PathCfsToOs` with `"internal//config.go"`.
Expect success — the path is normalized.

#### Rejects invalid CfsPath

Call `PathCfsToOs` with `"../../etc/passwd"`.
Expect error DirectoryTraversal.

#### Rejects symlink escaping project root

Create a symlink inside the project root pointing to a
directory outside it. Call `PathCfsToOs` with a path
through the symlink. Expect error ResolvesOutsideRoot.

#### Roundtrip: CfsToOs then OsToCfs

Call `PathCfsToOs` with `"internal/config/config.go"` to
get a `PathOs`. Then call `PathOsToCfs` with that result.
Expect the final `PathCfs` value equals
`"internal/config/config.go"`.

### PathOsToCfs

#### Converts valid OS path that exists

Given the project root, create a file inside it. Call
`PathOsToCfs` with the absolute OS path. Expect a
`PathCfs` with forward slashes, relative to the project
root.

#### Converts valid OS path that does not exist

Given the project root, construct an absolute OS path to
a file that does not exist but is within the root. Call
`PathOsToCfs`. Expect success — a `PathCfs` with forward
slashes.

#### Result uses forward slashes

On any OS, call `PathOsToCfs` with a valid absolute OS
path. Expect the resulting `PathCfs` contains no
backslashes.

#### Symlink within root resolving within root

Create a symlink inside the project root pointing to
another location inside the project root. Call
`PathOsToCfs` with the symlink path. Expect success.

#### Rejects path outside project root

Call `PathOsToCfs` with an absolute OS path that is
outside the project root. Expect error
ResolvesOutsideRoot.

### PathGetProjectRoot

#### Returns an absolute path

Call `PathGetProjectRoot`. Expect the result is a
`PathOs` that is a non-empty absolute path.

#### Matches working directory

Call `PathGetProjectRoot`. Expect the result corresponds
to the current working directory of the process.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function names from the interface:
  `PathValidateCfs`, `PathCfsToOs`, `PathOsToCfs`,
  `PathGetProjectRoot`.
