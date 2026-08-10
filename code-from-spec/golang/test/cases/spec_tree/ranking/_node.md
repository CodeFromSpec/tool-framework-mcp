---
depends_on:
  - SPEC/golang/test/utils/helpers
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/spec_tree/ranking
output: internal/noderanking/noderanking_test.go
---

# SPEC/golang/test/cases/spec_tree/ranking

# Agent

## Test cases

In v5, there is no bare "SPEC" root node. Root nodes
are direct children of code-from-spec/ (e.g.
"SPEC/root"). Tests use "SPEC/root" as the root node
where a tree hierarchy is needed.

### Happy path

#### Root only

Setup:
- entries = [parsing.Node with
  Reference.LogicalName = "SPEC/root",
  Reference.ParentName = nil, Frontmatter = nil]

Actions:
1. Call noderanking.NodeRankCompute(entries).

Expected: ranked = [{ Reference.LogicalName =
"SPEC/root", Rank = 0 }],
cycles = [].

#### Linear chain — incrementing ranks

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/a/b]
  (parent chain, no imports).

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

#### imports increases rank

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b
  where SPEC/root/b has
  imports = ["SPEC/root/a"]].

Expected: rank of SPEC/root/b > rank of SPEC/root/a.
cycles = [].

#### imports with qualifier — qualifier stripped

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b
  where SPEC/root/b has
  imports = ["SPEC/root/a(interface)"]].

Expected: No error. rank of SPEC/root/b >
rank of SPEC/root/a. cycles = [].

#### EXTERNAL imports — skipped for ranking

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["EXTERNAL/proto/api.proto"]].

Expected: No error. SPEC/root/a rank = 1. cycles = [].

#### input artifact adds dependency edge

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "out.go",
  SPEC/root/b with
  input = ["ARTIFACT/root/a"]].

Expected: rank of SPEC/root/b > rank of
ARTIFACT/root/a > rank of SPEC/root/a. cycles = [].

#### SPEC input adds dependency edge

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b with
  input = ["SPEC/root/a"]].

Expected: rank of SPEC/root/b > rank of SPEC/root/a.
cycles = [].

#### EXTERNAL input — skipped for ranking

Setup:
- entries = [SPEC/root, SPEC/root/a with
  input = ["EXTERNAL/docs/spec.yaml"]].

Expected: No error. SPEC/root/a rank = 1. cycles = [].

#### Multiple input entries — rank uses max

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "a.go",
  SPEC/root/b with type = "artifact" and
  output = "b.go",
  SPEC/root/c with
  input = ["ARTIFACT/root/a", "ARTIFACT/root/b"]].

Expected: rank of SPEC/root/c is strictly greater than
the rank of both ARTIFACT/root/a and ARTIFACT/root/b.
cycles = [].

#### Artifacts get rank one above their node

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "foo.go"].

Expected: ARTIFACT/root/a rank =
rank of SPEC/root/a + 1. cycles = [].

#### Single output — artifact ranked

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "x.go"].

Expected: ranked contains ARTIFACT/root/a with
rank = rank of SPEC/root/a + 1. cycles = [].

#### imports ARTIFACT reference — used as-is

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "lib.go",
  SPEC/root/b with
  imports = ["ARTIFACT/root/a"]].

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
  imports = ["SPEC/root/c"], SPEC/root/b with
  imports = ["SPEC/root/c"], SPEC/root/d with
  imports = ["SPEC/root/a", "SPEC/root/b"]].

Expected: SPEC/root/c=1, SPEC/root/a=2,
SPEC/root/b=2, SPEC/root/d=3. cycles = [].

#### imports outranks parent

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/a/b
  with imports = ["SPEC/root/c"], SPEC/root/c,
  SPEC/root/c/d, SPEC/root/c/d/e].

Expected: rank of SPEC/root/a/b > rank of SPEC/root/a.
SPEC/root/a/b rank = 1 + max(rank of SPEC/root/a,
rank of SPEC/root/c). cycles = [].

#### Multiple imports — rank from highest

Setup:
- entries = [SPEC/root, SPEC/root/a, SPEC/root/b with
  imports = ["SPEC/root/a"], SPEC/root/c with
  imports = ["SPEC/root/b"], SPEC/root/d with
  imports = ["SPEC/root/a", "SPEC/root/b",
  "SPEC/root/c"]].

Expected: SPEC/root/a=1, SPEC/root/b=2,
SPEC/root/c=3, SPEC/root/d=4. cycles = [].

#### Node with both imports and input

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "a.go",
  SPEC/root/b, SPEC/root/c with
  imports = ["SPEC/root/b"] and
  input = ["ARTIFACT/root/a"]].

Expected: rank of SPEC/root/c = 1 + max(rank of
SPEC/root, rank of SPEC/root/b,
rank of ARTIFACT/root/a). cycles = [].

#### Empty input list

Setup:
- entries = [].

Expected: ranked = [], cycles = [].

#### Verdict node produces VERDICT virtual entry

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "verdict"].

Expected: ranked contains VERDICT/root/a with
rank = rank of SPEC/root/a + 1. cycles = [].

#### Verdict node with default output produces virtual entry

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "verdict", no explicit output].

Expected: ranked contains VERDICT/root/a. No error.

#### Artifact node with default output produces virtual entry

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact", no explicit output].

Expected: ranked contains ARTIFACT/root/a. No error.

#### wait_on ARTIFACT raises rank

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "a.go",
  SPEC/root/b with type = "verdict" and
  wait_on = ["ARTIFACT/root/a"]].

Expected: rank of SPEC/root/b > rank of
ARTIFACT/root/a. cycles = [].

#### wait_on VERDICT raises rank

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "verdict",
  SPEC/root/b with type = "verdict" and
  wait_on = ["VERDICT/root/a"]].

Expected: rank of SPEC/root/b > rank of
VERDICT/root/a. cycles = [].

#### wait_on with ARTIFACT glob

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "verdict" and
  wait_on = ["ARTIFACT/root/b/*"],
  SPEC/root/b,
  SPEC/root/b/x with type = "artifact" and
  output = "x.go",
  SPEC/root/b/y with type = "artifact" and
  output = "y.go"].

Expected: SPEC/root/a rank > ARTIFACT/root/b/x rank
and > ARTIFACT/root/b/y rank.

#### wait_on cycle detected

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "a.go" and
  wait_on = ["ARTIFACT/root/b"],
  SPEC/root/b with type = "artifact" and
  output = "b.go" and
  wait_on = ["ARTIFACT/root/a"]].

Expected: cycles is not empty.

#### Unresolvable wait_on ARTIFACT reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "verdict" and
  wait_on = ["ARTIFACT/root/missing"]].

Expected: Error noderanking.ErrUnresolvableReference.

#### Unresolvable wait_on VERDICT reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "verdict" and
  wait_on = ["VERDICT/root/missing"]].

Expected: Error noderanking.ErrUnresolvableReference.

### Cycle detection

#### Self-reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/a"]].

Expected: cycles is not empty.

#### Simple cycle — two nodes

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/b"], SPEC/root/b with
  imports = ["SPEC/root/a"]].

Expected: cycles is not empty, contains at least one
of SPEC/root/a or SPEC/root/b.

#### Cycle through artifacts

Setup:
- entries = [SPEC/root, SPEC/root/a with
  type = "artifact" and output = "a.go" and
  imports = ["ARTIFACT/root/b"], SPEC/root/b with
  type = "artifact" and output = "b.go" and
  imports = ["ARTIFACT/root/a"]].

Expected: cycles is not empty.

#### Cycle does not prevent ranking of unrelated nodes

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/b"], SPEC/root/b with
  imports = ["SPEC/root/a"], SPEC/root/c
  (no deps)].

Expected: SPEC/root rank 0, SPEC/root/c rank 1.
cycles is not empty, contains entries related to
SPEC/root/a and/or SPEC/root/b but not SPEC/root/c.

### Glob expansion

#### SPEC glob creates dependency edges

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/b/*"],
  SPEC/root/b, SPEC/root/b/x, SPEC/root/b/y].

Expected: SPEC/root/a rank > SPEC/root/b/x rank
and > SPEC/root/b/y rank.

#### ARTIFACT glob creates dependency edges

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["ARTIFACT/root/b/*"],
  SPEC/root/b,
  SPEC/root/b/x with type = "artifact" and
  output = "x.go",
  SPEC/root/b/y with type = "artifact" and
  output = "y.go"].

Expected: SPEC/root/a rank > ARTIFACT/root/b/x rank
and > ARTIFACT/root/b/y rank.

#### Glob with empty match — no error

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/empty/*"],
  SPEC/root/empty].

Expected: No error. SPEC/root/a rank = 1
(parent dep only).

#### Cycle through glob

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/b/*"],
  SPEC/root/b with imports = ["SPEC/root/a"]].
  Note: SPEC/root/b has no descendants, so glob
  expands to nothing. No cycle.

Expected: No cycle. SPEC/root/a rank = 1.

### Error cases

#### Unresolvable SPEC reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["SPEC/root/missing"]].

Expected: Error noderanking.ErrUnresolvableReference.

#### Unresolvable ARTIFACT reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  imports = ["ARTIFACT/root/missing"]].

Expected: Error noderanking.ErrUnresolvableReference.

#### Unresolvable ARTIFACT input reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  input = ["ARTIFACT/root/missing"]].

Expected: Error noderanking.ErrUnresolvableReference.

#### Unresolvable SPEC input reference

Setup:
- entries = [SPEC/root, SPEC/root/a with
  input = ["SPEC/root/missing"]].

Expected: Error noderanking.ErrUnresolvableReference.

## Go-specific guidance

- The package name is `noderanking_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Build parsing.Node records directly — no file I/O.
- Set Reference.ParentName to nil for root nodes
  (e.g. "SPEC/root"), and to the parent logical name
  for nested nodes (e.g. ParentName = pointer to
  "SPEC/root" for "SPEC/root/a").
- Set Frontmatter to nil for nodes without frontmatter
  fields. Use *string for Input and Output.
