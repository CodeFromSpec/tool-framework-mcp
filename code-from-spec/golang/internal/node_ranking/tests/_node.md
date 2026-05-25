---
outputs:
  - id: noderanking_test
    path: internal/noderanking/noderanking_test.go
---

# ROOT/golang/internal/node_ranking/tests

Test cases for the noderanking package.

# Agent

## Context

Each test uses `t.TempDir()` to create an isolated temporary
directory. Create `_node.md` files with controlled frontmatter
to set up dependency graphs. Use table-driven tests where
appropriate.

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

### Artifacts get rank one above their node

Create ROOT/a with an output artifact. Expect the artifact
entry has rank = rank of ROOT/a + 1.

## Failure Cases

### Circular dependency detected

Create ROOT/a depends_on ROOT/b and ROOT/b depends_on
ROOT/a. Expect both logical names appear in the cycle
participants list.

### Unresolvable reference

Create a node with `depends_on` pointing to a non-existent
node. Expect `errors.Is(err, ErrUnresolvableRef)`.
