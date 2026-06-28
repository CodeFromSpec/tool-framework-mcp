---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/utils/logical_names/output.md
---

# SPEC/golang/interfaces/utils/logical_names

Maps logical names to file paths and provides utilities
for navigating the spec tree hierarchy.

# Public

## Package

`package logicalnames`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"`

## Interface

```go
func LogicalNameToPath(logicalName string) (pathutils.PathCfs, error)
func LogicalNameFromPath(cfsPath pathutils.PathCfs) (string, error)
func LogicalNameGetParent(logicalName string) (string, error)
func LogicalNameGetQualifier(logicalName string) (string, bool)
func LogicalNameStripQualifier(logicalName string) string
func LogicalNameHasParent(logicalName string) bool
func LogicalNameHasQualifier(logicalName string) bool
func LogicalNameIsArtifact(logicalName string) bool
func LogicalNameIsSpec(logicalName string) bool
func LogicalNameIsExternal(logicalName string) bool
func LogicalNameGetArtifactGenerator(logicalName string) (string, error)
func LogicalNameExternalToPath(logicalName string) (pathutils.PathCfs, error)
```

### Errors

- `ErrUnsupportedReference` (LogicalNameToPath)
- `ErrInvalidPath` (LogicalNameFromPath)
- `ErrNoParent` (LogicalNameGetParent)
- `ErrNotASpecReference` (LogicalNameGetParent)
- `ErrNotAnArtifactReference` (LogicalNameGetArtifactGenerator)
- `ErrNotAnExternalReference` (LogicalNameExternalToPath)

### Path resolution

| Logical name | PathCfs |
|---|---|
| `SPEC` | `code-from-spec/_node.md` |
| `SPEC/x/y` | `code-from-spec/x/y/_node.md` |
| `SPEC/x/y(z)` | `code-from-spec/x/y/_node.md` |

### Reverse resolution

| PathCfs | Logical name |
|---|---|
| `code-from-spec/_node.md` | `SPEC` |
| `code-from-spec/x/y/_node.md` | `SPEC/x/y` |

# Agent

Generate an interface specification document listing
the package, import path, function signatures, and
error sentinels.
