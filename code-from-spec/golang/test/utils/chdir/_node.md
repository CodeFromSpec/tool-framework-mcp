---
output: internal/testutils/chdir.go
---

# SPEC/golang/test/utils/chdir

Test helper that changes the working directory to a
temporary directory and restores it on cleanup.

# Public

## Package

`package testutils`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"`

```go
func Chdir(t *testing.T) string
```

Creates a temporary directory via `t.TempDir()`,
changes the working directory to it, and registers
a `t.Cleanup` to restore the original directory when
the test finishes. Returns the path to the temporary
directory.

Calls `t.Helper()` so that failures are reported at
the caller's line. Calls `t.Fatalf` if `os.Getwd` or
`os.Chdir` fails. Calls `t.Errorf` if the cleanup
restore fails.

# Agent

Implement the function listed in the Ownership section
as a Go file in package `testutils`.

## Ownership

This file declares and implements:
- Functions: `Chdir`

## Reference implementation

```go
func Chdir(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()
    orig, err := os.Getwd()
    if err != nil {
        t.Fatalf("Chdir: %v", err)
    }
    if err := os.Chdir(dir); err != nil {
        t.Fatalf("Chdir: %v", err)
    }
    t.Cleanup(func() {
        if err := os.Chdir(orig); err != nil {
            t.Errorf("Chdir cleanup: %v", err)
        }
    })
    return dir
}
```
