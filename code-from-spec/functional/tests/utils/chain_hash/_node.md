---
depends_on:
  - ROOT/functional/logic/utils/chain_hash(interface)
outputs:
  - id: chain_hash_tests
    path: code-from-spec/functional/tests/utils/chain_hash/output.md
---

# ROOT/functional/tests/utils/chain_hash

Test cases for the chain hash component.

# Public

## Test cases

### Happy path

#### Hash is deterministic

Create a spec tree with known content. Compute the chain
hash twice on the same tree. Expect both results are
identical.

#### Hash is 27 characters

Compute the chain hash for any valid spec tree. Expect the
result is exactly 27 characters long.

#### Hash changes when a file in the chain changes

Create a spec tree and compute the hash. Modify the content
of one file in the chain and recompute. Expect the two
hashes differ.

### Failure cases

#### Qualified depends_on with different case

Create a spec tree where node A has
`depends_on: ROOT/b(interface)` and node B has a
`# Public` section with a `## Interface` subsection
(capital I). Compute the chain hash for node A. Change
the content of `## Interface` in node B and recompute.
Expect the two hashes differ — the subsection must be
found regardless of case differences between the
qualifier and the heading.

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
