---
output: code-from-spec/golang/interfaces/os/path_utils/output.md
---

# SPEC/golang/interfaces/os/path_utils

Path types and safe path conversion for the framework.

# Public

## Package

`package pathutils`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"`

## Interface

```go
type PathCfs struct {
	Value string
}

type PathOs struct {
	Value string
}

func PathGetProjectRoot() (PathOs, error)
func PathValidateCfs(value string) error
func PathCfsToOs(cfsPath PathCfs) (PathOs, error)
func PathOsToCfs(osPath PathOs) (PathCfs, error)
```

### PathCfs

A path in the Code from Spec standard format:
- Forward slash as separator, always.
- Relative to the project root.
- No `..` components, no drive letters, no leading `/`,
  no backslashes.

### PathOs

An absolute path in the OS's native format. Never
exposed in the framework's public API.

### Errors

- `ErrCannotDetermineRoot` (PathGetProjectRoot)
- `ErrPathEmpty`, `ErrPathAbsolute`,
  `ErrPathContainsBackslash`, `ErrDirectoryTraversal`
  (PathValidateCfs)
- `ErrResolvesOutsideRoot` (PathCfsToOs, PathOsToCfs)
- Propagated from PathValidateCfs, PathGetProjectRoot.

# Agent

Generate an interface specification document listing
the package, import path, struct definitions, and
function signatures.
