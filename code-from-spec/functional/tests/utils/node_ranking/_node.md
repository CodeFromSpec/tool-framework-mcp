---
depends_on:
  - ROOT/functional/logic/utils/node_ranking(interface)
outputs:
  - id: node_ranking_tests
    path: code-from-spec/functional/tests/utils/node_ranking/output.md
---

# ROOT/functional/tests/utils/node_ranking

Test cases for the node ranking component.

# Public

## Test cases

### Happy path

#### Linear chain has incrementing ranks

Create three nodes: ROOT, ROOT/a, ROOT/a/b (parent chain).
Call DetectCycles. Expect ranks 0, 1, 2 respectively. No
cycle participants.

#### Independent siblings have equal rank

Create ROOT and two children ROOT/a and ROOT/b with no
cross-dependencies. Call DetectCycles. Expect ROOT/a and
ROOT/b have the same rank. No cycle participants.

#### depends_on increases rank

Create ROOT, ROOT/a, ROOT/b where ROOT/b depends_on
ROOT/a. Call DetectCycles. Expect ROOT/b has higher rank
than ROOT/a. No cycle participants.

#### depends_on with qualifier resolves correctly

Create ROOT, ROOT/a with a public section containing a
subsection, and ROOT/b with depends_on ROOT/a(interface).
Call DetectCycles. Expect no error -- the qualified
reference resolves to ROOT/a. Expect ROOT/b has higher
rank than ROOT/a. No cycle participants.

#### Artifacts get rank one above their node

Create ROOT/a with an output artifact. Call DetectCycles.
Expect the artifact entry has rank = rank of ROOT/a + 1.

### Failure cases

#### Circular dependency detected

Create ROOT/a depends_on ROOT/b and ROOT/b depends_on
ROOT/a. Call DetectCycles. Expect cycle participants list
is not empty.

#### Unresolvable reference

Create a node with depends_on pointing to a non-existent
node. Call DetectCycles. Expect "unresolvable reference".

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
