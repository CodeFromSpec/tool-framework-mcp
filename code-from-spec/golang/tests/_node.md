# ROOT/golang/tests

Go test files generated from functional test specifications
and implementation artifacts.

# Public

## Test rules

- Translate functional test cases from the `input`
  artifact into Go test functions.
- Use the interface and implementation from `depends_on`
  for types, function signatures, and import paths.
- Each test uses `t.TempDir()` for isolation.
- Create test files with controlled content using
  `os.WriteFile`.
- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.
- Prefix test helper functions with `test`.
