---
depends_on:
  - ROOT/functional/logic/mcp_tools/write_file(interface)
outputs:
  - id: write_file_tests
    path: code-from-spec/functional/tests/mcp_tools/write_file/output.md
---

# ROOT/functional/tests/mcp_tools/write_file

Test cases for the write file tool.

# Public

## Test cases

### Happy path

#### Writes file successfully

Create a spec tree with ROOT/a having outputs pointing to
output/file.go. Call HandleWriteFile with logical name =
"ROOT/a", path = "output/file.go", and content =
"package main".

Expect success with message "wrote output/file.go". Verify
the file exists on disk with the correct content.

#### Creates intermediate directories

Create a spec tree with ROOT/a having outputs pointing to
deep/nested/dir/file.go. Call HandleWriteFile with path =
"deep/nested/dir/file.go".

Expect success. Directories created automatically.

#### Overwrites existing file

Create a spec tree with ROOT/a having outputs pointing to
output/file.go. Write an initial file at that path. Call
HandleWriteFile with new content.

Expect success. File content replaced.

#### Path with backslashes is normalized (Windows only)

On Windows: create a spec tree with ROOT/a having outputs
pointing to output/file.go. Call HandleWriteFile with
path = "output\\file.go" and content = "package main".

Expect success with message "wrote output/file.go". The
backslash path matches the forward-slash outputs entry
after normalization.

### Failure cases

#### Invalid logical name prefix

Call HandleWriteFile with an invalid logical name. Expect
error.

#### Nonexistent logical name

Call HandleWriteFile with a logical name whose spec file
does not exist. Expect error.

#### Path not in outputs

Create a spec tree with ROOT/a having outputs pointing to
allowed/file.go. Call HandleWriteFile with path =
"other/file.go".

Expect error containing "path not allowed" and listing the
allowed paths.

#### Path traversal attempt

Create a spec tree with ROOT/a having outputs pointing to
"../../etc/passwd". Call HandleWriteFile with that path.

Expect error from path validation.

#### Empty path

Create a spec tree with ROOT/a having outputs pointing to
some/file.go. Call HandleWriteFile with path = "".

Expect error containing "path is empty".

#### Symlink escaping project root

Create a symlink inside the temporary directory pointing
outside it. Create a spec tree with the symlink path in
outputs. Call HandleWriteFile with that path.

Expect error containing "resolves outside project root".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
