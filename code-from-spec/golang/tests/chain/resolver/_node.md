---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/parsing(interface)
output: internal/chainresolver/chainresolver_test.go
---

# SPEC/golang/tests/chain/resolver

# Agent

## Test setup guidance

Tests create a spec tree on disk with `_node.md` files
containing frontmatter as needed, then call
`ChainResolve`. Use `testChdir` and create
`code-from-spec/.../_node.md` files.

In v5, there is no bare "SPEC" root node. Root nodes
are direct children of code-from-spec/. Tests use
"SPEC/root" as a root node where needed.

## Test cases

### Ancestors and target

#### Root as target

Setup:
- Create `code-from-spec/root/_node.md` with empty
  frontmatter.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root").

Expected:
- ancestors = empty list.
- dependencies = empty list.
- target = parsing.CfsReference("SPEC/root", Qualifier=nil).
- input = absent.

#### Linear chain — ancestors in root-first order

Setup:
- Create SPEC/root, SPEC/root/a, SPEC/root/a/b with
  empty frontmatter.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a/b").

Expected:
- ancestors = [SPEC/root, SPEC/root/a] in that order.
- target = SPEC/root/a/b.

#### Single parent

Setup:
- Create SPEC/root, SPEC/root/a with empty frontmatter.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- ancestors = [SPEC/root].
- target = SPEC/root/a.

#### Target with empty frontmatter

Setup:
- Create SPEC/root, SPEC/root/a with empty frontmatter.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- ancestors = [SPEC/root], target = SPEC/root/a,
  dependencies = empty, input = absent.

### Dependencies — SPEC/ references

#### Dependency without qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b"]), SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- dependencies contains one parsing.CfsReference with
  LogicalName = "SPEC/root/b",
  Qualifier = nil.

#### Dependency with qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b(interface)"]),
  SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- dependencies contains one parsing.CfsReference with
  LogicalName = "SPEC/root/b",
  Qualifier = "interface".

#### Dependencies sorted by logical name

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/z", "SPEC/root/m",
  "SPEC/root/b"]), SPEC/root/z, SPEC/root/m,
  SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- dependencies = [SPEC/root/b, SPEC/root/m,
  SPEC/root/z] in that order.

### Dependencies — ARTIFACT/ references

#### ARTIFACT dependency resolved from generating node

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["ARTIFACT/root/b"]),
  SPEC/root/b (output = "out/lib.go").

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- dependencies contains one parsing.CfsReference with
  LogicalName = "ARTIFACT/root/b",
  Path = "out/lib.go".

#### ARTIFACT — generating node has no output

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["ARTIFACT/root/b"]),
  SPEC/root/b (empty frontmatter, no output).

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: Error chainresolver.ErrUnresolvableArtifact.

#### ARTIFACT — artifact file does not exist on disk

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["ARTIFACT/root/b"]),
  SPEC/root/b (output = "out/lib.go").
- Do NOT create "out/lib.go" on disk.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- No error. dependencies contains one CfsReference
  with Path = "out/lib.go". Existence is not
  verified.

#### Mixed SPEC/, ARTIFACT/, EXTERNAL/ dependencies

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/c", "ARTIFACT/root/b",
  "EXTERNAL/proto/api.proto"]),
  SPEC/root/b (output = "out/lib.go"), SPEC/root/c.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected:
- dependencies sorted: ARTIFACT/root/b,
  EXTERNAL/proto/api.proto, SPEC/root/c.

### Dependencies — dedup

#### Exact duplicate — same file, same qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b", "SPEC/root/b"]),
  SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains one entry for
SPEC/root/b.

#### No qualifier subsumes qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b",
  "SPEC/root/b(interface)"]), SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains one entry for
SPEC/root/b with Qualifier = nil.

#### Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b(interface)",
  "SPEC/root/b"]), SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains one entry for
SPEC/root/b with Qualifier = nil.

#### Same file, different qualifiers — both kept

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b(interface)",
  "SPEC/root/b(constraints)"]), SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains two entries, one
with Qualifier = "constraints", one with "interface"
(sorted).

#### Duplicate ARTIFACT — same logical name

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["ARTIFACT/root/b",
  "ARTIFACT/root/b"]),
  SPEC/root/b (output = "out/lib.go").

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains one ARTIFACT entry.

### Dependencies — EXTERNAL/ references

#### EXTERNAL dependency resolved to path

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["EXTERNAL/docs/api.yaml"]).

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains one parsing.CfsReference with
LogicalName = "EXTERNAL/docs/api.yaml",
Path = "docs/api.yaml", Qualifier = nil.

#### Multiple EXTERNAL dependencies sorted

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["EXTERNAL/proto/v1.proto",
  "EXTERNAL/docs/api.yaml"]).

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies sorted:
EXTERNAL/docs/api.yaml, EXTERNAL/proto/v1.proto.

#### Duplicate EXTERNAL — same logical name

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["EXTERNAL/x.proto",
  "EXTERNAL/x.proto"]).

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: dependencies contains one EXTERNAL entry.

### Input

#### Input resolved from generating node

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "ARTIFACT/root/b"),
  SPEC/root/b (output = "out/data.json").

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: input = parsing.CfsReference("ARTIFACT/root/b",
Path = "out/data.json").

#### EXTERNAL input resolved to path

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "EXTERNAL/docs/vendor/spec.yaml").

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: input = parsing.CfsReference(
"EXTERNAL/docs/vendor/spec.yaml",
Path = "docs/vendor/spec.yaml").

#### SPEC input resolved

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "SPEC/root/b"), SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: input = parsing.CfsReference("SPEC/root/b",
Path = "code-from-spec/root/b/_node.md",
Qualifier = nil).

#### SPEC input with qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "SPEC/root/b(acceptance-tests)"),
  SPEC/root/b.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: input = parsing.CfsReference("SPEC/root/b",
Path = "code-from-spec/root/b/_node.md",
Qualifier = "acceptance-tests").

#### No input — absent

Setup:
- Create SPEC/root, SPEC/root/a (no input).

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: input is absent (nil).

### Error cases

#### Unrecognized prefix in depends_on

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["UNKNOWN/something"]).

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: Error chainresolver.ErrUnresolvableArtifact.

#### Invalid target logical name

Actions:
1. Call chainresolver.ChainResolve("INVALID/something").

Expected: Error propagated from
parsing.CfsReferenceFromName.

#### Input ARTIFACT — generating node not found

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "ARTIFACT/root/missing").
- Do not create SPEC/root/missing.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: Error propagated from
parsing.CfsReferenceFromName (generator node missing).

#### Unreadable frontmatter

Setup:
- Create SPEC/root, SPEC/root/a with invalid YAML
  in frontmatter.

Actions:
1. Call chainresolver.ChainResolve("SPEC/root/a").

Expected: Error chainresolver.ErrUnreadableFrontmatter.

## Go-specific guidance

- The package name is `chainresolver_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
