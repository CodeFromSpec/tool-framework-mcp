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
