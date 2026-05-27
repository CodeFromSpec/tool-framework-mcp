---
depends_on:
  - ROOT/functional/logic/utils/node_discovery(interface)
outputs:
  - id: node_discovery_tests
    path: code-from-spec/functional/tests/utils/node_discovery/output.md
---

# ROOT/functional/tests/utils/node_discovery

Test cases for the node discovery component.

# Public

## Test cases

### Happy path

#### Discovers nodes in a simple tree

Create a code-from-spec directory with a root node file and
one sub-node file. Call DiscoverNodes. Expect two entries,
sorted alphabetically by logical name.

#### Ignores non-node files

Create a code-from-spec directory with a root node file and
a non-node file (e.g. README.md) in a subdirectory. Call
DiscoverNodes. Expect only one entry for the root node.

#### Result is sorted by logical name

Create several nodes at different depths. Call
DiscoverNodes. Expect the returned list is sorted
alphabetically by logical name.

### Failure cases

#### No code-from-spec directory

Do not create a code-from-spec directory. Call
DiscoverNodes. Expect "directory not found".

#### Empty code-from-spec directory

Create a code-from-spec directory but no node files inside.
Call DiscoverNodes. Expect "no nodes found".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
