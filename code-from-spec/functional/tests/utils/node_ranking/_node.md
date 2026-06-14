---
depends_on:
  - SPEC/functional/logic/utils/node_ranking(interface)
output: code-from-spec/functional/tests/utils/node_ranking/output.md
---

# SPEC/functional/tests/utils/node_ranking

Test cases for the node ranking component.

# Public

## Test cases

### Happy path

#### Root only

Input: single entry with logical_name = "SPEC", empty
frontmatter. Call NodeRankCompute.

Expect one NodeRankEntry: "SPEC" with rank 0. No cycles.

#### Linear chain — incrementing ranks

Input: SPEC, SPEC/a, SPEC/a/b (parent chain, no
depends_on). Call NodeRankCompute.

Expect ranks: SPEC=0, SPEC/a=1, SPEC/a/b=2. No cycles.

#### Independent siblings — equal rank

Input: SPEC, SPEC/a, SPEC/b (no cross-dependencies).
Call NodeRankCompute.

Expect SPEC/a and SPEC/b have the same rank (1). No
cycles.

#### depends_on increases rank

Input: SPEC, SPEC/a, SPEC/b where SPEC/b has
depends_on = ["SPEC/a"]. Call NodeRankCompute.

Expect SPEC/b has higher rank than SPEC/a. No cycles.

#### depends_on with qualifier — qualifier stripped

Input: SPEC, SPEC/a, SPEC/b where SPEC/b has
depends_on = ["SPEC/a(interface)"]. Call NodeRankCompute.

Expect no error. SPEC/b has higher rank than SPEC/a.
The qualified reference resolves to the bare node
SPEC/a. No cycles.

#### EXTERNAL depends_on — skipped for ranking

Input: SPEC, SPEC/a with depends_on =
["EXTERNAL/proto/api.proto"]. Call NodeRankCompute.

Expect no error. SPEC/a rank = 1 (parent only).
EXTERNAL/ references do not contribute to rank.
No cycles.

#### input artifact adds dependency edge

Input: SPEC, SPEC/a with output = "out.go",
SPEC/b with input = "ARTIFACT/a".
Call NodeRankCompute.

Expect SPEC/b has higher rank than the artifact
ARTIFACT/a, which has higher rank than SPEC/a.
No cycles.

#### EXTERNAL input — skipped for ranking

Input: SPEC, SPEC/a with input =
"EXTERNAL/docs/spec.yaml". Call NodeRankCompute.

Expect no error. SPEC/a rank = 1 (parent only).
EXTERNAL/ input does not contribute to rank. No cycles.

#### Artifacts get rank one above their node

Input: SPEC, SPEC/a with output = "foo.go".
Call NodeRankCompute.

Expect ARTIFACT/a has rank = rank of SPEC/a + 1.
No cycles.

#### Single output — artifact ranked

Input: SPEC, SPEC/a with output = "x.go".
Call NodeRankCompute.

Expect one artifact entry ARTIFACT/a with rank =
rank of SPEC/a + 1. No cycles.

#### depends_on ARTIFACT reference — used as-is

Input: SPEC, SPEC/a with output = "lib.go",
SPEC/b with depends_on = ["ARTIFACT/a"].
Call NodeRankCompute.

Expect SPEC/b rank > ARTIFACT/a rank > SPEC/a
rank. No cycles.

#### Output sorted by rank then logical name

Input: SPEC, SPEC/z, SPEC/a (no cross-dependencies).
Call NodeRankCompute.

Expect output order: SPEC (rank 0), then SPEC/a before
SPEC/z (both rank 1, alphabetical). No cycles.

#### Parallel entries — equal rank means no dependency

Input: SPEC, SPEC/a, SPEC/b, SPEC/c — all siblings,
no cross-dependencies. Call NodeRankCompute.

Expect SPEC/a, SPEC/b, SPEC/c all have rank 1. No
cycles.

#### Diamond dependency — rank uses max not sum

Input: SPEC, SPEC/c, SPEC/a with depends_on =
["SPEC/c"], SPEC/b with depends_on = ["SPEC/c"],
SPEC/d with depends_on = ["SPEC/a", "SPEC/b"].
Call NodeRankCompute.

Expect SPEC/c=1, SPEC/a=2, SPEC/b=2, SPEC/d=3 (not
5). Rank is 1 + max of dependencies, not sum. No
cycles.

#### depends_on outranks parent

Input: SPEC, SPEC/a, SPEC/a/b with depends_on =
["SPEC/c"], SPEC/c, SPEC/c/d, SPEC/c/d/e (deep chain).
Call NodeRankCompute.

Expect SPEC/a/b rank > SPEC/a rank, because depends_on
SPEC/c contributes rank. SPEC/a/b rank = 1 + max(rank
of SPEC/a, rank of SPEC/c). No cycles.

#### Multiple depends_on — rank from highest

Input: SPEC, SPEC/a, SPEC/b, SPEC/c, SPEC/d with
depends_on = ["SPEC/a", "SPEC/b", "SPEC/c"] where
SPEC/b has depends_on = ["SPEC/a"] and SPEC/c has
depends_on = ["SPEC/b"]. Call NodeRankCompute.

Expect SPEC/a=1, SPEC/b=2, SPEC/c=3, SPEC/d=4 (1 +
max(1,2,3) = 4, not 1 + first or last). No cycles.

#### Node with both depends_on and input

Input: SPEC, SPEC/a with output = "a.go",
SPEC/b, SPEC/c with depends_on = ["SPEC/b"] and
input = "ARTIFACT/a". Call NodeRankCompute.

Expect SPEC/c rank = 1 + max(rank of SPEC (parent),
rank of SPEC/b, rank of ARTIFACT/a). No cycles.

#### Empty input list

Input: empty list. Call NodeRankCompute.

Expect empty ranked list, no cycles.

### Cycle detection

#### Self-reference

Input: SPEC, SPEC/a with depends_on = ["SPEC/a"].
Call NodeRankCompute.

Expect cycles list is not empty.

#### Simple cycle — two nodes

Input: SPEC, SPEC/a with depends_on = ["SPEC/b"],
SPEC/b with depends_on = ["SPEC/a"].
Call NodeRankCompute.

Expect cycles list is not empty and contains at least
one of SPEC/a or SPEC/b.

#### Cycle through artifacts

Input: SPEC, SPEC/a with output = "a.go" and
depends_on = ["ARTIFACT/b"], SPEC/b with
output = "b.go" and depends_on = ["ARTIFACT/a"].
Call NodeRankCompute.

Expect cycles list is not empty.

#### Cycle does not prevent ranking of unrelated nodes

Input: SPEC, SPEC/a with depends_on = ["SPEC/b"],
SPEC/b with depends_on = ["SPEC/a"], SPEC/c (no
dependencies). Call NodeRankCompute.

Expect SPEC and SPEC/c have valid ranks (0 and 1).
Cycles list contains entries related to SPEC/a and/or
SPEC/b but not SPEC/c.

### Error cases

#### Unresolvable SPEC reference

Input: SPEC, SPEC/a with depends_on = ["SPEC/missing"].
Call NodeRankCompute.

Expect error UnresolvableReference.

#### Unresolvable ARTIFACT reference

Input: SPEC, SPEC/a with depends_on =
["ARTIFACT/missing"]. Call NodeRankCompute.

Expect error UnresolvableReference.

#### Unresolvable input reference

Input: SPEC, SPEC/a with input =
"ARTIFACT/missing". Call NodeRankCompute.

Expect error UnresolvableReference.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `NodeRankCompute`.
- Use the record names from the interface: `NodeRankInput`,
  `NodeRankEntry`.
- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (input data),
  actions (function call), and expected outcome.
- Input is always a list of `NodeRankInput` — no file I/O
  in tests.
