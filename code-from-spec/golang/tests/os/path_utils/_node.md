---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/implementation/os/path_utils(pathutils)
input: ARTIFACT/functional/tests/os/path_utils(path_utils_tests)
outputs:
  - id: pathutils_test
    path: internal/pathutils/pathutils_test.go
---

# ROOT/golang/tests/os/path_utils

Go test file for the `pathutils` package, generated from
the functional test specification.

# Agent

Translate the functional test cases from the input into
a Go test file for the `pathutils` package.

## Go-specific guidance

- Each test uses `t.TempDir()` as the project root.
- When testing symlinks, create them inside the temp
  directory pointing to targets outside it.
- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.
- Prefix test helper functions with `test`.
