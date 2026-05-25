---
outputs:
  - id: nodediscovery_test
    path: internal/nodediscovery/nodediscovery_test.go
---

# ROOT/golang/internal/node_discovery/tests

Test cases for the nodediscovery package.

# Agent

## Context

Each test uses `t.TempDir()` to create an isolated temporary
directory and `os.Chdir` to set it as the working directory
(restore the original directory with `t.Cleanup`). Create a
`code-from-spec/` directory structure with `_node.md` files
inside the temp directory.

## Happy Path

### Discovers nodes in a simple tree

Create `code-from-spec/_node.md` and
`code-from-spec/sub/_node.md`. Call `DiscoverNodes`.
Expect two entries, sorted alphabetically by logical name.

### Ignores non-node files

Create `code-from-spec/_node.md` and
`code-from-spec/sub/README.md`. Call `DiscoverNodes`.
Expect only one entry for the root node.

### Result is sorted by logical name

Create several nodes at different depths. Verify the
returned slice is sorted alphabetically by `LogicalName`.

## Failure Cases

### No code-from-spec directory

Do not create `code-from-spec/`. Call `DiscoverNodes`.
Expect `errors.Is(err, ErrDirNotFound)`.

### Empty code-from-spec directory

Create `code-from-spec/` but no `_node.md` files inside.
Call `DiscoverNodes`.
Expect `errors.Is(err, ErrNoNodesFound)`.
