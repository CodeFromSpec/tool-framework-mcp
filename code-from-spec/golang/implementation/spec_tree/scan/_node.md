---
depends_on:
  - SPEC/golang/implementation/os/list_files
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/utils/logical_names
output: internal/spectree/spectree.go
---

# SPEC/golang/implementation/spec_tree/scan

Scans the `code-from-spec/` directory and returns all
spec nodes found.

# Public

## Package

`package spectree`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectree"`

## Interface

```go
type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}

func SpecTreeScan() ([]*SpecTreeNode, error)
```

Takes no parameters. Scans the `code-from-spec/`
directory relative to the project root. Returns a
list sorted alphabetically by logical name.

### Errors

- `ErrNoNodesFound`: no `_node.md` files found under
  `code-from-spec/`.
- Propagated errors from `listfiles`, `logicalnames`
  packages.

# Agent

Implement the spec tree scan as a Go package. The output
file is the sole .go file in the package — declare all
types, error sentinels, and function signatures from the
interface artifact in this file.

## Logic

1. Call `ListFiles` with "code-from-spec/" as the
   directory. If `ListFiles` raises an error, propagate
   it.

2. Filter the list: keep only files whose name after
   the last "/" is exactly "_node.md".

3. For each remaining file, exclude it if:
   a. It is directly inside "code-from-spec/" (i.e.
      `code-from-spec/_node.md`). There is no root
      node — only subdirectories are nodes.
   b. It lives inside a `.`-prefixed directory directly
      under "code-from-spec/":
        Remove the "code-from-spec/" prefix from the
        file path. Look for the first "/" in the
        remainder. If a "/" is found, extract the text
        before it as the first directory segment. If
        the first directory segment starts with ".",
        exclude this file.

4. For each file that was not excluded, call
   `LogicalNameFromPath` with the file's PathCfs.
   If `LogicalNameFromPath` raises an error, propagate
   it. Let `ln` be the result. Build a SpecTreeNode
   record with: logical_name = ln.Name,
   file_path = the file's PathCfs.

5. Sort all resulting SpecTreeNode records alphabetically
   by logical_name.

6. If the sorted list is empty, raise error
   "no nodes found".

7. Return the sorted list of SpecTreeNode records.

## Go-specific guidance

- Use the `listfiles` package for `ListFiles`.
- Use the `logicalnames` package for `LogicalNameFromPath`.
- Use the `pathutils` package for `PathCfs`.
- Extract the file name by finding the last `/` in the
  `PathCfs.Value` string.
- The package name should be `spectree`.
