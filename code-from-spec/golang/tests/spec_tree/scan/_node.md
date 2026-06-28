---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/scan
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/list_files
  - ARTIFACT/golang/interfaces/utils/logical_names
output: internal/spectree/spectree_test.go
---

# SPEC/golang/tests/spec_tree/scan

# Agent

## Test cases

### Happy path

#### Root node only

Setup:
- Create `code-from-spec/_node.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected:
- One SpecTreeNode with LogicalName = `"SPEC"` and
  FilePath = `code-from-spec/_node.md`.

#### Root and nested nodes

Setup:
- Create `code-from-spec/_node.md`,
  `code-from-spec/a/_node.md`, and
  `code-from-spec/a/b/_node.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected:
- Three entries: SPEC, SPEC/a, SPEC/a/b with correct
  file paths.

#### Ignores non-node files

Setup:
- Create `code-from-spec/_node.md` and
  `code-from-spec/x/output.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Only one entry for SPEC.

#### Ignores _-prefixed directories under code-from-spec

Setup:
- Create `code-from-spec/_node.md`,
  `code-from-spec/_rules/some/_node.md`, and
  `code-from-spec/_tools/_node.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Only one entry for SPEC.

#### _-prefixed dirs deeper in tree are NOT ignored

Setup:
- Create `code-from-spec/_node.md`,
  `code-from-spec/a/_node.md`, and
  `code-from-spec/a/_internal/_node.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Three entries: SPEC, SPEC/a,
SPEC/a/_internal.

#### Ignores directories without _node.md

Setup:
- Create `code-from-spec/_node.md` and an empty
  subdirectory `code-from-spec/x/y/`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Only one entry for SPEC.

#### Result is sorted by logical name

Setup:
- Create nodes at `code-from-spec/z/_node.md`,
  `code-from-spec/_node.md`,
  `code-from-spec/a/b/_node.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Sorted order: SPEC, SPEC/a/b, SPEC/z.

### Failure cases

#### No code-from-spec directory

Setup:
- Do not create `code-from-spec/`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Error propagated from ListFiles.

#### Empty code-from-spec directory

Setup:
- Create `code-from-spec/` with no files.

Actions:
1. Call `SpecTreeScan()`.

Expected: Error `ErrNoNodesFound`.

#### Only non-node files in code-from-spec

Setup:
- Create `code-from-spec/README.md` and
  `code-from-spec/x/output.md` but no `_node.md`.

Actions:
1. Call `SpecTreeScan()`.

Expected: Error `ErrNoNodesFound`.

## Go-specific guidance

- The package name is `spectree_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
