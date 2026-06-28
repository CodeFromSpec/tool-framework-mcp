---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
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
1. Call ChainResolve("SPEC/root").

Expected:
- ancestors = empty list.
- dependencies = empty list.
- target = ChainItem("SPEC/root", qualifier=absent).
- input = absent.

#### Linear chain — ancestors in root-first order

Setup:
- Create SPEC/root, SPEC/root/a, SPEC/root/a/b with
  empty frontmatter.

Actions:
1. Call ChainResolve("SPEC/root/a/b").

Expected:
- ancestors = [SPEC/root, SPEC/root/a] in that order.
- target = SPEC/root/a/b.

#### Single parent

Setup:
- Create SPEC/root, SPEC/root/a with empty frontmatter.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected:
- ancestors = [SPEC/root].
- target = SPEC/root/a.

#### Target with empty frontmatter

Setup:
- Create SPEC/root, SPEC/root/a with empty frontmatter.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected:
- ancestors = [SPEC/root], target = SPEC/root/a,
  dependencies = empty, input = absent.

### Dependencies — SPEC/ references

#### Dependency without qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b"]), SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected:
- dependencies contains one ChainItem with
  UnqualifiedLogicalName = "SPEC/root/b",
  Qualifier = absent.

#### Dependency with qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b(interface)"]),
  SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected:
- dependencies contains one ChainItem with
  UnqualifiedLogicalName = "SPEC/root/b",
  Qualifier = "interface".

#### Dependencies sorted by logical name

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/z", "SPEC/root/m",
  "SPEC/root/b"]), SPEC/root/z, SPEC/root/m,
  SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

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
1. Call ChainResolve("SPEC/root/a").

Expected:
- dependencies contains one ChainItem with
  UnqualifiedLogicalName = "ARTIFACT/root/b",
  FilePath = "out/lib.go".

#### ARTIFACT — generating node has no output

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["ARTIFACT/root/b"]),
  SPEC/root/b (empty frontmatter, no output).

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: Error ErrUnresolvableArtifact.

#### ARTIFACT — artifact file does not exist on disk

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["ARTIFACT/root/b"]),
  SPEC/root/b (output = "out/lib.go").
- Do NOT create "out/lib.go" on disk.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected:
- No error. dependencies contains one ChainItem
  with FilePath = "out/lib.go". Existence is not
  verified.

#### Mixed SPEC/, ARTIFACT/, EXTERNAL/ dependencies

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/c", "ARTIFACT/root/b",
  "EXTERNAL/proto/api.proto"]),
  SPEC/root/b (output = "out/lib.go"), SPEC/root/c.

Actions:
1. Call ChainResolve("SPEC/root/a").

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
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies contains one entry for
SPEC/root/b.

#### No qualifier subsumes qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b",
  "SPEC/root/b(interface)"]), SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies contains one entry for
SPEC/root/b with Qualifier = absent.

#### Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b(interface)",
  "SPEC/root/b"]), SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies contains one entry for
SPEC/root/b with Qualifier = absent.

#### Same file, different qualifiers — both kept

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["SPEC/root/b(interface)",
  "SPEC/root/b(constraints)"]), SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

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
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies contains one ARTIFACT entry.

### Dependencies — EXTERNAL/ references

#### EXTERNAL dependency resolved to path

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["EXTERNAL/docs/api.yaml"]).

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies contains one ChainItem with
UnqualifiedLogicalName = "EXTERNAL/docs/api.yaml",
FilePath = "docs/api.yaml", Qualifier = absent.

#### Multiple EXTERNAL dependencies sorted

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["EXTERNAL/proto/v1.proto",
  "EXTERNAL/docs/api.yaml"]).

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies sorted:
EXTERNAL/docs/api.yaml, EXTERNAL/proto/v1.proto.

#### Duplicate EXTERNAL — same logical name

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["EXTERNAL/x.proto",
  "EXTERNAL/x.proto"]).

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: dependencies contains one EXTERNAL entry.

### Input

#### Input resolved from generating node

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "ARTIFACT/root/b"),
  SPEC/root/b (output = "out/data.json").

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: input = ChainItem("ARTIFACT/root/b",
FilePath = "out/data.json").

#### EXTERNAL input resolved to path

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "EXTERNAL/docs/vendor/spec.yaml").

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: input = ChainItem(
"EXTERNAL/docs/vendor/spec.yaml",
FilePath = "docs/vendor/spec.yaml").

#### SPEC input resolved

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "SPEC/root/b"), SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: input = ChainItem("SPEC/root/b",
FilePath = "code-from-spec/root/b/_node.md",
Qualifier = absent).

#### SPEC input with qualifier

Setup:
- Create SPEC/root, SPEC/root/a
  (input = "SPEC/root/b(acceptance-tests)"),
  SPEC/root/b.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: input = ChainItem("SPEC/root/b",
FilePath = "code-from-spec/root/b/_node.md",
Qualifier = "acceptance-tests").

#### No input — absent

Setup:
- Create SPEC/root, SPEC/root/a (no input).

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: input is absent (nil).

### Error cases

#### Unrecognized prefix in depends_on

Setup:
- Create SPEC/root, SPEC/root/a
  (depends_on = ["UNKNOWN/something"]).

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: Error ErrUnresolvableArtifact.

#### Invalid target logical name

Actions:
1. Call ChainResolve("INVALID/something").

Expected: Error propagated from LogicalNameParse.

#### Unreadable frontmatter

Setup:
- Create SPEC/root, SPEC/root/a with invalid YAML
  in frontmatter.

Actions:
1. Call ChainResolve("SPEC/root/a").

Expected: Error ErrUnreadableFrontmatter.

## Go-specific guidance

- The package name is `chainresolver_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
