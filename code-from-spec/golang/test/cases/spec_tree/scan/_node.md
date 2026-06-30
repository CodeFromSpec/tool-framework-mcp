---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/spec_tree/scan
  - SPEC/golang/implementation/parsing(interface)
output: internal/spectree/spectree_test.go
---

# SPEC/golang/test/cases/spec_tree/scan

# Agent

## Test cases

### Happy path

#### Single root node

Setup:
- Create `code-from-spec/a/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected:
- One `parsing.CfsReference` with LogicalName = `"SPEC/a"` and
  Path = `"code-from-spec/a/_node.md"`.

#### Multiple root nodes

Setup:
- Create `code-from-spec/a/_node.md` and
  `code-from-spec/b/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected:
- Two entries: SPEC/a and SPEC/b.

#### Root and nested nodes

Setup:
- Create `code-from-spec/a/_node.md` and
  `code-from-spec/a/b/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected:
- Two entries: SPEC/a, SPEC/a/b with correct
  file paths.

#### Ignores non-node files

Setup:
- Create `code-from-spec/a/_node.md` and
  `code-from-spec/x/output.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Only one entry for SPEC/a.

#### Ignores .-prefixed directories under code-from-spec

Setup:
- Create `code-from-spec/a/_node.md`,
  `code-from-spec/.cache/some/_node.md`, and
  `code-from-spec/.hidden/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Only one entry for SPEC/a.

#### .-prefixed dirs deeper in tree are ignored

Setup:
- Create `code-from-spec/a/_node.md` and
  `code-from-spec/a/.internal/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Only one entry for SPEC/a. The node under
the `.`-prefixed directory is excluded.

#### Ignores _node.md directly in code-from-spec/

Setup:
- Create `code-from-spec/_node.md` and
  `code-from-spec/a/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Only one entry for SPEC/a. The root
`code-from-spec/_node.md` is excluded.

#### Ignores directories without _node.md

Setup:
- Create `code-from-spec/a/_node.md` and an empty
  subdirectory `code-from-spec/x/y/`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Only one entry for SPEC/a.

#### Result is sorted by logical name

Setup:
- Create nodes at `code-from-spec/z/_node.md`,
  `code-from-spec/a/_node.md`,
  `code-from-spec/a/b/_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Sorted order: SPEC/a, SPEC/a/b, SPEC/z.

### Failure cases

#### No code-from-spec directory

Setup:
- Do not create `code-from-spec/`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: `oslayer.ErrDirectoryNotFound` propagated
from `oslayer.ListAllFiles`.

#### Empty code-from-spec directory

Setup:
- Create `code-from-spec/` with no files.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Error `spectree.ErrNoNodesFound`.

#### Only non-node files in code-from-spec

Setup:
- Create `code-from-spec/README.md` and
  `code-from-spec/x/output.md` but no `_node.md`.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Error `spectree.ErrNoNodesFound`.

#### Only root _node.md — no subdirectory nodes

Setup:
- Create `code-from-spec/_node.md` but no subdirectory
  nodes.

Actions:
1. Call `spectree.SpecTreeScan()`.

Expected: Error `spectree.ErrNoNodesFound`.

## Go-specific guidance

- The package name is `spectree_test` (external test
  package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
