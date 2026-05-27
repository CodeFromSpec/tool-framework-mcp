---
depends_on:
  - ROOT/functional/logic/utils/path_validation(interface)
outputs:
  - id: path_validation_tests
    path: code-from-spec/functional/tests/utils/path_validation/output.md
---

# ROOT/functional/tests/utils/path_validation

Test cases for the path validation component.

# Public

## Test cases

### Happy path

#### Simple relative path

Input: "internal/config/config.go", project root = temporary
directory. Call ValidatePath. Expect no error.

#### Nested path

Input: "cmd/framework-mcp/main.go", project root = temporary
directory. Call ValidatePath. Expect no error.

#### Single filename

Input: "main.go", project root = temporary directory. Call
ValidatePath. Expect no error.

#### Path with dot segment

Input: "internal/./config/config.go", project root =
temporary directory. Call ValidatePath. Expect no error
(cleaned to "internal/config/config.go").

### Edge cases

#### Path with trailing slash

Input: "internal/config/", project root = temporary
directory. Call ValidatePath. Expect no error.

#### Path with duplicate separators

Input: "internal//config//config.go", project root =
temporary directory. Call ValidatePath. Expect no error
(cleaned by path normalization).

### Failure cases

#### Empty path

Input: "", project root = temporary directory. Call
ValidatePath. Expect error containing "path is empty".

#### Absolute path with leading slash

Input: "/etc/passwd", project root = temporary directory.
Call ValidatePath. Expect error containing "path is
absolute".

#### Absolute path with drive letter (Windows-style)

Input: "C:\\Windows\\system32", project root = temporary
directory. Call ValidatePath. Expect error containing "path
is absolute".

#### Simple traversal

Input: "../../etc/passwd", project root = temporary
directory. Call ValidatePath. Expect error containing
"directory traversal".

#### Embedded traversal

Input: "internal/../../outside/file.go", project root =
temporary directory. Call ValidatePath. Expect error
containing "directory traversal".

#### Symlink escaping project root

Create a symlink inside the temporary directory pointing to
a directory outside it. Input: "<symlink>/file.go", project
root = temporary directory. Call ValidatePath. Expect error
containing "resolves outside project root".

#### Traversal disguised with dot segments

Input: "a/../../outside", project root = temporary
directory. Call ValidatePath. Expect error containing
"directory traversal".

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
