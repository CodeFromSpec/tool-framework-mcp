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
Expect one `SpecTreeNode` with `logical_name` = `"SPEC"`
and `file_path` = `code-from-spec/_node.md`.

#### Root and nested nodes

Create `code-from-spec/_node.md`,
`code-from-spec/a/_node.md`, and
`code-from-spec/a/b/_node.md`. Call `SpecTreeScan`.
Expect three entries with correct logical names
(`SPEC`, `SPEC/a`, `SPEC/a/b`) and correct file paths.

#### Ignores non-node files

Create `code-from-spec/_node.md` and
`code-from-spec/x/output.md` (not a `_node.md`).
Call `SpecTreeScan`. Expect only one entry for `SPEC`.

#### Ignores _-prefixed directories under code-from-spec

Create `code-from-spec/_node.md`,
`code-from-spec/_rules/some/_node.md`, and
`code-from-spec/_tools/_node.md`. Call `SpecTreeScan`.
Expect only one entry for `SPEC` — nodes inside
`_rules/` and `_tools/` are ignored.

#### _-prefixed dirs deeper in tree are NOT ignored

Create `code-from-spec/_node.md`,
`code-from-spec/a/_node.md`, and
`code-from-spec/a/_internal/_node.md`. Call
`SpecTreeScan`. Expect three entries: `SPEC`, `SPEC/a`,
`SPEC/a/_internal` — the `_` prefix only applies to
directories directly under `code-from-spec/`.

#### Ignores directories without _node.md

Create `code-from-spec/_node.md` and an empty
subdirectory `code-from-spec/x/y/`. Call
`SpecTreeScan`. Expect only one entry for `SPEC`.

#### Result is sorted by logical name

Create nodes at various depths in non-alphabetical
order (e.g., `code-from-spec/z/_node.md`,
`code-from-spec/_node.md`,
`code-from-spec/a/b/_node.md`). Call `SpecTreeScan`.
Expect the returned list is sorted alphabetically by
logical name: `SPEC`, `SPEC/a/b`, `SPEC/z`.

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
