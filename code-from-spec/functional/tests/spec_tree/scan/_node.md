---
depends_on:
  - ROOT/functional/logic/spec_tree/scan(interface)
output: code-from-spec/functional/tests/spec_tree/scan/output.md
---

# ROOT/functional/tests/spec_tree/scan

Test cases for the spec tree scanning component.

# Public

## Test cases

### Happy path

#### Root node only

Create `code-from-spec/_node.md`. Call `SpecTreeScan`.
Expect one `SpecTreeNode` with `logical_name` = `"ROOT"`
and `file_path` = `code-from-spec/_node.md`.

#### Root and nested nodes

Create `code-from-spec/_node.md`,
`code-from-spec/a/_node.md`, and
`code-from-spec/a/b/_node.md`. Call `SpecTreeScan`.
Expect three entries with correct logical names
(`ROOT`, `ROOT/a`, `ROOT/a/b`) and correct file paths.

#### Ignores non-node files

Create `code-from-spec/_node.md` and
`code-from-spec/x/output.md` (not a `_node.md`).
Call `SpecTreeScan`. Expect only one entry for `ROOT`.

#### Ignores directories without _node.md

Create `code-from-spec/_node.md` and an empty
subdirectory `code-from-spec/x/y/`. Call
`SpecTreeScan`. Expect only one entry for `ROOT`.

#### Result is sorted by logical name

Create nodes at various depths in non-alphabetical
order (e.g., `code-from-spec/z/_node.md`,
`code-from-spec/_node.md`,
`code-from-spec/a/b/_node.md`). Call `SpecTreeScan`.
Expect the returned list is sorted alphabetically by
logical name: `ROOT`, `ROOT/a/b`, `ROOT/z`.

### Failure cases

#### No code-from-spec directory

Do not create a `code-from-spec/` directory. Call
`SpecTreeScan`. Expect error propagated from
`ListFiles` (directory not found).

#### Empty code-from-spec directory

Create `code-from-spec/` with no files inside. Call
`SpecTreeScan`. Expect error NoNodesFound.

#### Only non-node files in code-from-spec

Create `code-from-spec/README.md` and
`code-from-spec/x/output.md` but no `_node.md` files.
Call `SpecTreeScan`. Expect error NoNodesFound.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `SpecTreeScan`.
