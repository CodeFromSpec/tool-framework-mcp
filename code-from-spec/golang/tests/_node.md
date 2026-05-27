# ROOT/golang/tests

Go test files generated from functional test specifications
and implementation artifacts.

# Public

## Test rules

- Translate functional test cases from the `input`
  artifact into Go test functions.
- Use the interface from `depends_on` for types, function
  signatures, and import paths.
- Use the external test package (`package <name>_test`)
  for black-box testing. Import the package under test
  explicitly. This avoids name collisions and ensures
  tests exercise the public API only.
- Each test uses `t.TempDir()` for isolation.
- Create test files with controlled content using
  `os.WriteFile`.
- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.
