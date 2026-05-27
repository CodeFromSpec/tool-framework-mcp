# ROOT/golang/tests/os

Go tests for OS abstraction components.

# Public

## Test environment

Components in the `os/` layer resolve paths against the
project root, which is the working directory of the
process. Tests must control this:

1. Create a temporary directory with `t.TempDir()`.
2. Change the working directory to it with `os.Chdir`.
3. Restore the original working directory in `t.Cleanup`.
4. Create test files inside this temporary directory.
5. Pass `PathCfs` values that are relative to the
   temporary directory (which is now the working
   directory).

This ensures that `PathCfsToOs` and other path-resolving
functions find the test files correctly.
