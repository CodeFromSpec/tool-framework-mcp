# ROOT/tech_design/internal/pathvalidation

Validates that a file path is safe to write to within a project
directory. This is a security-critical package — it prevents
writing files outside the intended project boundary.

# Public

## Package

`package pathvalidation`

## Threat model

When a tool accepts a file path as input and writes to disk,
the path could attempt to escape the project directory using:

- **Relative traversal**: `../../etc/passwd`
- **Embedded traversal**: `internal/../../outside/file.go`
- **OS-specific separators**: backslash on Windows (`..\..\`)
- **Encoding tricks**: URL-encoded or Unicode sequences
- **Symlinks**: a valid relative path that resolves outside
  the project via a symlink in the directory tree

This package provides a single validation function that callers
use before any write operation.

## Interface

```go
func ValidatePath(path string, projectRoot string) error
```

Returns nil if the path is safe to write within `projectRoot`.
Returns an error describing the violation otherwise.

### Error messages

- `"path is empty"`
- `"path is absolute: <path>"`
- `"path contains directory traversal: <path>"`
- `"path resolves outside project root: <path>"`
