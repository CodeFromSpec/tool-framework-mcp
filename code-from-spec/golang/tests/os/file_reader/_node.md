---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/implementation/os/file_reader(filereader)
input: ARTIFACT/functional/tests/os/file_reader(file_reader_tests)
outputs:
  - id: filereader_test
    path: internal/filereader/filereader_test.go
---

# ROOT/golang/tests/os/file_reader

Go test file for the filereader package, generated from
the functional test specification.

# Agent

Translate the functional test cases from the input into
a Go test file for the `filereader` package.

## Go-specific guidance

- Each test uses `t.TempDir()` to create an isolated
  temporary directory.
- Create test files with controlled content using
  `os.WriteFile`.
- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.
- Prefix test helper functions with `test`.
