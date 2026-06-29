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

In v5, there is no bare "SPEC" root node. Root nodes
are direct children of code-from-spec/ (e.g.
"SPEC/root"). Tests use "SPEC/root" as the root node
where a tree hierarchy is needed.

### Happy path

#### Root only

Setup:
- entries = [NodeRankInput { LogicalName: "SPEC/root",
  Parent: nil, Frontmatter: empty }]

Actions:
1. Call NodeRankCompute(entries).

Expected: ranked = [{ "SPEC/root", rank: 0 }],
cycles = [].

#### Linear chain — incrementing ranks

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/a/b]
  (parent chain, no depends_on).

Expected: SPEC/root=0, SPEC/root/a=1,
SPEC/root/a/b=2. cycles = [].

#### Independent siblings — equal rank

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b]
  (no cross-deps).

Expected: SPEC/root/a and SPEC/root/b both rank 1.
cycles = [].

#### Multiple independent roots

Setup:
- entries = [SPEC/alpha, SPEC/beta] (two independent
  root nodes, no cross-deps).

Expected: SPEC/alpha=0, SPEC/beta=0. cycles = [].

#### depends_on increases rank

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b
  where SPEC/root/b has
  depends_on = ["SPEC/root/a"]].

Expected: rank of SPEC/root/b > rank of SPEC/root/a.
cycles = [].

#### depends_on with qualifier — qualifier stripped

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b
  where SPEC/root/b has
  depends_on = ["SPEC/root/a(interface)"]].

Expected: No error. rank of SPEC/root/b >
rank of SPEC/root/a. cycles = [].

#### EXTERNAL depends_on — skipped for ranking

Setup:
- entries = [SPEC/root, SPEC/root/a with
  depends_on = ["EXTERNAL/proto/api.proto"]].

Expected: No error. SPEC/root/a rank = 1. cycles = [].

#### input artifact adds dependency edge

Setup:
- entries = [SPEC/root, SPEC/root/a with
  output = "out.go", SPEC/root/b with
  input = "ARTIFACT/root/a"].

Expected: rank of SPEC/root/b > rank of
ARTIFACT/root/a > rank of SPEC/root/a. cycles = [].

#### EXTERNAL input — skipped for ranking

Setup:
- entries = [SPEC/root, SPEC/root/a with
  input = "EXTERNAL/docs/spec.yaml"].

Expected: No error. SPEC/root/a rank = 1. cycles = [].

#### Artifacts get rank one above their node

Setup:
- entries = [SPEC/root, SPEC/root/a with
  output = "foo.go"].

Expected: ARTIFACT/root/a rank =
rank of SPEC/root/a + 1. cycles = [].

#### Single output — artifact ranked

Setup:
- entries = [SPEC/root, SPEC/root/a with
  output = "x.go"].

Expected: ranked contains ARTIFACT/root/a with
rank = rank of SPEC/root/a + 1. cycles = [].

#### depends_on ARTIFACT reference — used as-is

Setup:
- entries = [SPEC/root, SPEC/root/a with
  output = "lib.go", SPEC/root/b with
  depends_on = ["ARTIFACT/root/a"]].

Expected: rank of SPEC/root/b >
rank of ARTIFACT/root/a > rank of SPEC/root/a.
cycles = [].

#### Output sorted by rank then logical name

Setup:
- entries = [SPEC/root, SPEC/root/z, SPEC/root/a]
  (no cross-deps).

Expected: ranked[0] = SPEC/root (rank 0), then
SPEC/root/a before SPEC/root/z (both rank 1,
alphabetical). cycles = [].

#### Parallel entries — equal rank means no dependency

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b,
  SPEC/root/c] (all siblings, no cross-deps).

Expected: SPEC/root/a, SPEC/root/b, SPEC/root/c all
rank 1. cycles = [].

#### Diamond dependency — rank uses max not sum

Setup:
- entries = [SPEC/root, SPEC/root/c, SPEC/root/a with
  depends_on = ["SPEC/root/c"], SPEC/root/b with
  depends_on = ["SPEC/root/c"], SPEC/root/d with
  depends_on = ["SPEC/root/a", "SPEC/root/b"]].

Expected: SPEC/root/c=1, SPEC/root/a=2,
SPEC/root/b=2, SPEC/root/d=3. cycles = [].

#### depends_on outranks parent

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/a/b
  with depends_on = ["SPEC/root/c"], SPEC/root/c,
  SPEC/root/c/d, SPEC/root/c/d/e].

Expected: rank of SPEC/root/a/b > rank of SPEC/root/a.
SPEC/root/a/b rank = 1 + max(rank of SPEC/root/a,
rank of SPEC/root/c). cycles = [].

#### Multiple depends_on — rank from highest

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b with
  depends_on = ["SPEC/root/a"], SPEC/root/c with
  depends_on = ["SPEC/root/b"], SPEC/root/d with
  depends_on = ["SPEC/root/a", "SPEC/root/b",
  "SPEC/root/c"]].

Expected: SPEC/root/a=1, SPEC/root/b=2,
SPEC/root/c=3, SPEC/root/d=4. cycles = [].

#### Node with both depends_on and input

Setup:
- entries = [SPEC/root, SPEC/root/a with
  output = "a.go", SPEC/root/b, SPEC/root/c with
  depends_on = ["SPEC/root/b"] and
  input = "ARTIFACT/root/a"].

Expected: rank of SPEC/root/c = 1 + max(rank of
SPEC/root, rank of SPEC/root/b,
rank of ARTIFACT/root/a). cycles = [].

#### Empty input list

Setup:
- entries = [].

Expected: ranked = [], cycles = [].

### Cycle detection

#### Self-reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  depends_on = ["SPEC/root/a"]].

Expected: cycles is not empty.

#### Simple cycle — two nodes

Setup:
- entries = [SPEC/root, SPEC/root/a with
  depends_on = ["SPEC/root/b"], SPEC/root/b with
  depends_on = ["SPEC/root/a"]].

Expected: cycles is not empty, contains at least one
of SPEC/root/a or SPEC/root/b.

#### Cycle through artifacts

Setup:
- entries = [SPEC/root, SPEC/root/a with
  output = "a.go" and
  depends_on = ["ARTIFACT/root/b"], SPEC/root/b with
  output = "b.go" and
  depends_on = ["ARTIFACT/root/a"]].

Expected: cycles is not empty.

#### Cycle does not prevent ranking of unrelated nodes

Setup:
- entries = [SPEC/root, SPEC/root/a with
  depends_on = ["SPEC/root/b"], SPEC/root/b with
  depends_on = ["SPEC/root/a"], SPEC/root/c
  (no deps)].

Expected: SPEC/root rank 0, SPEC/root/c rank 1.
cycles is not empty, contains entries related to
SPEC/root/a and/or SPEC/root/b but not SPEC/root/c.

### Error cases

#### Unresolvable SPEC reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  depends_on = ["SPEC/root/missing"]].

Expected: Error ErrUnresolvableReference.

#### Unresolvable ARTIFACT reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  depends_on = ["ARTIFACT/root/missing"]].

Expected: Error ErrUnresolvableReference.

#### Unresolvable input reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  input = "ARTIFACT/root/missing"].

Expected: Error ErrUnresolvableReference.

## Go-specific guidance

- The package name is `noderanking_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Build NodeRankInput records directly — no file I/O.
- Set Parent to nil for root nodes (e.g. "SPEC/root"),
  and to the parent logical name for nested nodes
  (e.g. Parent = pointer to "SPEC/root" for
  "SPEC/root/a").
