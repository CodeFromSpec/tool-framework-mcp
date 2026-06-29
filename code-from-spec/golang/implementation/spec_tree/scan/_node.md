---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
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
func SpecTreeScan() ([]parsing.CfsReference, error)
```

Takes no parameters. Scans the `code-from-spec/`
directory relative to the project root. Returns a
list sorted alphabetically by logical name.

### Errors

- `ErrNoNodesFound`: no `_node.md` files found under
  `code-from-spec/`.
- Propagated errors from `oslayer`, `parsing` packages.

# Agent

Implement the spec tree scan as a Go package.

## Logic

1. Call `ListAllFiles` with "code-from-spec/" as the
   directory. If `ListAllFiles` raises an error, propagate
   it.

2. Filter the list: keep only files whose name after
   the last "/" is exactly "_node.md".

3. For each remaining file, exclude it if:
   a. It is directly inside "code-from-spec/" (i.e.
      `code-from-spec/_node.md`). There is no root
      node — only subdirectories are nodes.
   b. Any segment of the path between "code-from-spec/"
      and the file name starts with ".":
        Remove the "code-from-spec/" prefix from the
        file path. Split the remainder by "/". Discard
        the last element (the file name). For each
        remaining segment, if the segment starts with
        ".", exclude this file.

4. For each file that was not excluded, call
   `CfsReferenceFromPath` with the file's CfsPath.
   If `CfsReferenceFromPath` raises an error, propagate
   it. Collect the returned `*CfsReference`.

5. Sort all results alphabetically by `LogicalName`.

6. If the sorted list is empty, raise ErrNoNodesFound.

7. Return the sorted list.

## Go-specific guidance

- Use the `oslayer` package for `ListAllFiles` and
  `CfsPath`.
- Use the `parsing` package for `CfsReferenceFromPath`.
- Extract the file name by finding the last `/` in the
  `CfsPath` string value.
- The package name should be `spectree`.
