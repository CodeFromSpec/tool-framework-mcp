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

## Temporary files and PathCfs

Tests that create files and pass them as `PathCfs` values
must use the `testChdir` pattern:

1. Create a temp dir with `t.TempDir()`.
2. Call `os.Chdir(tempDir)` to make it the working
   directory. Register `t.Cleanup` to restore the
   original directory.
3. Create files using paths relative to the temp dir
   (e.g., `os.WriteFile("mydir/file.txt", ...)`).
4. Pass those relative paths as `PathCfs.Value`.

This works because `PathGetProjectRoot` returns the
working directory, and `PathCfsToOs` resolves relative
paths against it. Without `testChdir`, `t.TempDir()`
creates directories in the OS temp location, which may
be on a different drive (Windows) or outside the project
root — causing path resolution to fail.

A typical `testChdir` helper:

```go
func testChdir(t *testing.T, dir string) {
    t.Helper()
    orig, err := os.Getwd()
    if err != nil {
        t.Fatalf("testChdir: %v", err)
    }
    if err := os.Chdir(dir); err != nil {
        t.Fatalf("testChdir: %v", err)
    }
    t.Cleanup(func() {
        if err := os.Chdir(orig); err != nil {
            t.Errorf("testChdir cleanup: %v", err)
        }
    })
}
```

Tests that do not create files (pure function tests)
do not need this pattern.

## Error and style conventions

- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.
