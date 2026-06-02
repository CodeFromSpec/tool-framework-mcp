---
depends_on:
  - ROOT/functional/logic/utils/node_ranking(interface)
output: code-from-spec/functional/tests/utils/node_ranking/output.md
---

# ROOT/functional/tests/utils/node_ranking

Test cases for the node ranking component.

# Public

## Test cases

### Happy path

#### Root only

Input: single entry with logical_name = "ROOT", empty
frontmatter. Call NodeRankCompute.

Expect one NodeRankEntry: "ROOT" with rank 0. No cycles.

#### Linear chain — incrementing ranks

Input: ROOT, ROOT/a, ROOT/a/b (parent chain, no
depends_on). Call NodeRankCompute.

Expect ranks: ROOT=0, ROOT/a=1, ROOT/a/b=2. No cycles.

#### Independent siblings — equal rank

Input: ROOT, ROOT/a, ROOT/b (no cross-dependencies).
Call NodeRankCompute.

Expect ROOT/a and ROOT/b have the same rank (1). No
cycles.

#### depends_on increases rank

Input: ROOT, ROOT/a, ROOT/b where ROOT/b has
depends_on = ["ROOT/a"]. Call NodeRankCompute.

Expect ROOT/b has higher rank than ROOT/a. No cycles.

#### depends_on with qualifier — qualifier stripped

Input: ROOT, ROOT/a, ROOT/b where ROOT/b has
depends_on = ["ROOT/a(interface)"]. Call NodeRankCompute.

Expect no error. ROOT/b has higher rank than ROOT/a.
The qualified reference resolves to the bare node
ROOT/a. No cycles.

#### input artifact adds dependency edge

Input: ROOT, ROOT/a with outputs = [{id: "code",
path: "out.go"}], ROOT/b with input =
"ARTIFACT/a(code)". Call NodeRankCompute.

Expect ROOT/b has higher rank than the artifact
ARTIFACT/a(code), which has higher rank than ROOT/a.
No cycles.

#### Artifacts get rank one above their node

Input: ROOT, ROOT/a with outputs = [{id: "foo",
path: "foo.go"}]. Call NodeRankCompute.

Expect ARTIFACT/a(foo) has rank = rank of ROOT/a + 1.
No cycles.

#### Multiple outputs — each artifact ranked

Input: ROOT, ROOT/a with outputs = [{id: "x",
path: "x.go"}, {id: "y", path: "y.go"}].
Call NodeRankCompute.

Expect two artifact entries ARTIFACT/a(x) and
ARTIFACT/a(y), both with rank = rank of ROOT/a + 1.
No cycles.

#### depends_on ARTIFACT reference — used as-is

Input: ROOT, ROOT/a with outputs = [{id: "lib",
path: "lib.go"}], ROOT/b with depends_on =
["ARTIFACT/a(lib)"]. Call NodeRankCompute.

Expect ROOT/b rank > ARTIFACT/a(lib) rank > ROOT/a
rank. No cycles.

#### Output sorted by rank then logical name

Input: ROOT, ROOT/z, ROOT/a (no cross-dependencies).
Call NodeRankCompute.

Expect output order: ROOT (rank 0), then ROOT/a before
ROOT/z (both rank 1, alphabetical). No cycles.

#### Parallel entries — equal rank means no dependency

Input: ROOT, ROOT/a, ROOT/b, ROOT/c — all siblings,
no cross-dependencies. Call NodeRankCompute.

Expect ROOT/a, ROOT/b, ROOT/c all have rank 1. No
cycles.

#### Diamond dependency — rank uses max not sum

Input: ROOT, ROOT/c, ROOT/a with depends_on =
["ROOT/c"], ROOT/b with depends_on = ["ROOT/c"],
ROOT/d with depends_on = ["ROOT/a", "ROOT/b"].
Call NodeRankCompute.

Expect ROOT/c=1, ROOT/a=2, ROOT/b=2, ROOT/d=3 (not
5). Rank is 1 + max of dependencies, not sum. No
cycles.

#### depends_on outranks parent

Input: ROOT, ROOT/a, ROOT/a/b with depends_on =
["ROOT/c"], ROOT/c, ROOT/c/d, ROOT/c/d/e (deep chain).
Call NodeRankCompute.

Expect ROOT/a/b rank > ROOT/a rank, because depends_on
ROOT/c contributes rank. ROOT/a/b rank = 1 + max(rank
of ROOT/a, rank of ROOT/c). No cycles.

#### Multiple depends_on — rank from highest

Input: ROOT, ROOT/a, ROOT/b, ROOT/c, ROOT/d with
depends_on = ["ROOT/a", "ROOT/b", "ROOT/c"] where
ROOT/b has depends_on = ["ROOT/a"] and ROOT/c has
depends_on = ["ROOT/b"]. Call NodeRankCompute.

Expect ROOT/a=1, ROOT/b=2, ROOT/c=3, ROOT/d=4 (1 +
max(1,2,3) = 4, not 1 + first or last). No cycles.

#### Node with both depends_on and input

Input: ROOT, ROOT/a with outputs = [{id: "out",
path: "a.go"}], ROOT/b, ROOT/c with depends_on =
["ROOT/b"] and input = "ARTIFACT/a(out)".
Call NodeRankCompute.

Expect ROOT/c rank = 1 + max(rank of ROOT (parent),
rank of ROOT/b, rank of ARTIFACT/a(out)). No cycles.

#### Empty input list

Input: empty list. Call NodeRankCompute.

Expect empty ranked list, no cycles.

### Cycle detection

#### Self-reference

Input: ROOT, ROOT/a with depends_on = ["ROOT/a"].
Call NodeRankCompute.

Expect cycles list is not empty.

#### Simple cycle — two nodes

Input: ROOT, ROOT/a with depends_on = ["ROOT/b"],
ROOT/b with depends_on = ["ROOT/a"].
Call NodeRankCompute.

Expect cycles list is not empty and contains at least
one of ROOT/a or ROOT/b.

#### Cycle through artifacts

Input: ROOT, ROOT/a with outputs = [{id: "out",
path: "a.go"}] and depends_on = ["ARTIFACT/b(out)"],
ROOT/b with outputs = [{id: "out", path: "b.go"}]
and depends_on = ["ARTIFACT/a(out)"].
Call NodeRankCompute.

Expect cycles list is not empty.

#### Cycle does not prevent ranking of unrelated nodes

Input: ROOT, ROOT/a with depends_on = ["ROOT/b"],
ROOT/b with depends_on = ["ROOT/a"], ROOT/c (no
dependencies). Call NodeRankCompute.

Expect ROOT and ROOT/c have valid ranks (0 and 1).
Cycles list contains entries related to ROOT/a and/or
ROOT/b but not ROOT/c.

### Error cases

#### Unresolvable ROOT reference

Input: ROOT, ROOT/a with depends_on = ["ROOT/missing"].
Call NodeRankCompute.

Expect error UnresolvableReference.

#### Unresolvable ARTIFACT reference

Input: ROOT, ROOT/a with depends_on =
["ARTIFACT/missing(id)"]. Call NodeRankCompute.

Expect error UnresolvableReference.

#### Unresolvable input reference

Input: ROOT, ROOT/a with input =
"ARTIFACT/missing(id)". Call NodeRankCompute.

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
