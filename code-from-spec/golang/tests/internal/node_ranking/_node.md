---
depends_on:
  - ROOT/golang/implementation/internal/node_discovery
  - ROOT/golang/implementation/internal/frontmatter
input: ARTIFACT/golang/implementation/internal/node_ranking/code(noderanking)
outputs:
  - id: noderanking_test
    path: internal/noderanking/noderanking_test.go
---

# ROOT/golang/tests/internal/node_ranking

Test cases for the noderanking package.

# Agent

## Context

`DetectCycles` receives `[]nodediscovery.DiscoveredNode`
(with `LogicalName` and `FilePath` only — no frontmatter
field) and parses frontmatter internally from each
`FilePath`. Tests must therefore create real `_node.md`
files on disk with controlled frontmatter YAML content.

Each test uses `t.TempDir()` to create an isolated temporary
directory. Build `_node.md` files under it using a helper
that writes YAML frontmatter between `---` delimiters.
Construct `DiscoveredNode` slices manually — no need to call
`DiscoverNodes`.

Use table-driven tests where appropriate.

## Happy Path

### Linear chain has incrementing ranks

Create three nodes: ROOT, ROOT/a, ROOT/a/b (parent chain).
Expect ranks 0, 1, 2 respectively. No cycle participants.

### Independent siblings have equal rank

Create ROOT and two children ROOT/a and ROOT/b with no
cross-dependencies. Expect ROOT/a and ROOT/b have the same
rank. No cycle participants.

### depends_on increases rank

Create ROOT, ROOT/a, ROOT/b where ROOT/b depends_on ROOT/a.
Expect ROOT/b has higher rank than ROOT/a. No cycle
participants.

### depends_on with ROOT qualifier resolves correctly

Create ROOT, ROOT/a with a `# Public` section containing a
`## Interface` subsection, and ROOT/b with
`depends_on: ROOT/a(interface)`. Expect no error — the
qualified reference resolves to ROOT/a. Expect ROOT/b has
higher rank than ROOT/a. No cycle participants.

### Artifacts get rank one above their node

Create ROOT/a with an output artifact. Expect the artifact
entry has rank = rank of ROOT/a + 1.

## Failure Cases

### Circular dependency detected

Create ROOT/a depends_on ROOT/b and ROOT/b depends_on
ROOT/a. Expect cycle participants list is not empty.

### Unresolvable reference

Create a node with `depends_on` pointing to a non-existent
node. Expect `errors.Is(err, ErrUnresolvableRef)`.
