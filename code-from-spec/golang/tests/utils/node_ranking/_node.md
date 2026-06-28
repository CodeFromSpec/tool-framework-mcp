---
depends_on:
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
  - SPEC/golang/implementation/utils/node_ranking
output: internal/noderanking/noderanking_test.go
---

# SPEC/golang/tests/utils/node_ranking

# Agent

## Test cases

### Happy path

#### Root only

Setup:
- entries = [NodeRankInput { LogicalName: "SPEC",
  Frontmatter: empty }]

Actions:
1. Call NodeRankCompute(entries).

Expected: ranked = [{ "SPEC", rank: 0 }], cycles = [].

#### Linear chain — incrementing ranks

Setup:
- entries = [SPEC, SPEC/a, SPEC/a/b] (parent chain,
  no depends_on).

Expected: SPEC=0, SPEC/a=1, SPEC/a/b=2. cycles = [].

#### Independent siblings — equal rank

Setup:
- entries = [SPEC, SPEC/a, SPEC/b] (no cross-deps).

Expected: SPEC/a and SPEC/b both rank 1. cycles = [].

#### depends_on increases rank

Setup:
- entries = [SPEC, SPEC/a, SPEC/b where SPEC/b has
  depends_on = ["SPEC/a"]].

Expected: rank of SPEC/b > rank of SPEC/a.
cycles = [].

#### depends_on with qualifier — qualifier stripped

Setup:
- entries = [SPEC, SPEC/a, SPEC/b where SPEC/b has
  depends_on = ["SPEC/a(interface)"]].

Expected: No error. rank of SPEC/b > rank of SPEC/a.
cycles = [].

#### EXTERNAL depends_on — skipped for ranking

Setup:
- entries = [SPEC, SPEC/a with
  depends_on = ["EXTERNAL/proto/api.proto"]].

Expected: No error. SPEC/a rank = 1. cycles = [].

#### input artifact adds dependency edge

Setup:
- entries = [SPEC, SPEC/a with output = "out.go",
  SPEC/b with input = "ARTIFACT/a"].

Expected: rank of SPEC/b > rank of ARTIFACT/a >
rank of SPEC/a. cycles = [].

#### EXTERNAL input — skipped for ranking

Setup:
- entries = [SPEC, SPEC/a with
  input = "EXTERNAL/docs/spec.yaml"].

Expected: No error. SPEC/a rank = 1. cycles = [].

#### Artifacts get rank one above their node

Setup:
- entries = [SPEC, SPEC/a with output = "foo.go"].

Expected: ARTIFACT/a rank = rank of SPEC/a + 1.
cycles = [].

#### Single output — artifact ranked

Setup:
- entries = [SPEC, SPEC/a with output = "x.go"].

Expected: ranked contains ARTIFACT/a with
rank = rank of SPEC/a + 1. cycles = [].

#### depends_on ARTIFACT reference — used as-is

Setup:
- entries = [SPEC, SPEC/a with output = "lib.go",
  SPEC/b with depends_on = ["ARTIFACT/a"]].

Expected: rank of SPEC/b > rank of ARTIFACT/a >
rank of SPEC/a. cycles = [].

#### Output sorted by rank then logical name

Setup:
- entries = [SPEC, SPEC/z, SPEC/a] (no cross-deps).

Expected: ranked[0] = SPEC (rank 0), then SPEC/a
before SPEC/z (both rank 1, alphabetical). cycles = [].

#### Parallel entries — equal rank means no dependency

Setup:
- entries = [SPEC, SPEC/a, SPEC/b, SPEC/c] (all
  siblings, no cross-deps).

Expected: SPEC/a, SPEC/b, SPEC/c all rank 1.
cycles = [].

#### Diamond dependency — rank uses max not sum

Setup:
- entries = [SPEC, SPEC/c, SPEC/a with
  depends_on = ["SPEC/c"], SPEC/b with
  depends_on = ["SPEC/c"], SPEC/d with
  depends_on = ["SPEC/a", "SPEC/b"]].

Expected: SPEC/c=1, SPEC/a=2, SPEC/b=2, SPEC/d=3.
cycles = [].

#### depends_on outranks parent

Setup:
- entries = [SPEC, SPEC/a, SPEC/a/b with
  depends_on = ["SPEC/c"], SPEC/c, SPEC/c/d,
  SPEC/c/d/e].

Expected: rank of SPEC/a/b > rank of SPEC/a.
SPEC/a/b rank = 1 + max(rank of SPEC/a,
rank of SPEC/c). cycles = [].

#### Multiple depends_on — rank from highest

Setup:
- entries = [SPEC, SPEC/a, SPEC/b with
  depends_on = ["SPEC/a"], SPEC/c with
  depends_on = ["SPEC/b"], SPEC/d with
  depends_on = ["SPEC/a", "SPEC/b", "SPEC/c"]].

Expected: SPEC/a=1, SPEC/b=2, SPEC/c=3, SPEC/d=4.
cycles = [].

#### Node with both depends_on and input

Setup:
- entries = [SPEC, SPEC/a with output = "a.go",
  SPEC/b, SPEC/c with depends_on = ["SPEC/b"] and
  input = "ARTIFACT/a"].

Expected: rank of SPEC/c = 1 + max(rank of SPEC,
rank of SPEC/b, rank of ARTIFACT/a). cycles = [].

#### Empty input list

Setup:
- entries = [].

Expected: ranked = [], cycles = [].

### Cycle detection

#### Self-reference

Setup:
- entries = [SPEC, SPEC/a with
  depends_on = ["SPEC/a"]].

Expected: cycles is not empty.

#### Simple cycle — two nodes

Setup:
- entries = [SPEC, SPEC/a with
  depends_on = ["SPEC/b"], SPEC/b with
  depends_on = ["SPEC/a"]].

Expected: cycles is not empty, contains at least one
of SPEC/a or SPEC/b.

#### Cycle through artifacts

Setup:
- entries = [SPEC, SPEC/a with output = "a.go" and
  depends_on = ["ARTIFACT/b"], SPEC/b with
  output = "b.go" and depends_on = ["ARTIFACT/a"]].

Expected: cycles is not empty.

#### Cycle does not prevent ranking of unrelated nodes

Setup:
- entries = [SPEC, SPEC/a with
  depends_on = ["SPEC/b"], SPEC/b with
  depends_on = ["SPEC/a"], SPEC/c (no deps)].

Expected: SPEC rank 0, SPEC/c rank 1. cycles is not
empty, contains entries related to SPEC/a and/or
SPEC/b but not SPEC/c.

### Error cases

#### Unresolvable SPEC reference

Setup:
- entries = [SPEC, SPEC/a with
  depends_on = ["SPEC/missing"]].

Expected: Error ErrUnresolvableReference.

#### Unresolvable ARTIFACT reference

Setup:
- entries = [SPEC, SPEC/a with
  depends_on = ["ARTIFACT/missing"]].

Expected: Error ErrUnresolvableReference.

#### Unresolvable input reference

Setup:
- entries = [SPEC, SPEC/a with
  input = "ARTIFACT/missing"].

Expected: Error ErrUnresolvableReference.

## Go-specific guidance

- The package name is `noderanking_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Build NodeRankInput records directly — no file I/O.
