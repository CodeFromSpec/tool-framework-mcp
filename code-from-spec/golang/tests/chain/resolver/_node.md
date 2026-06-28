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

## Test cases

### Ancestors and target

#### Root as target

Setup:
- Create `code-from-spec/_node.md` with empty
  frontmatter.

Actions:
1. Call ChainResolve("SPEC").

Expected:
- ancestors = empty list.
- dependencies = empty list.
- target = ChainItem("SPEC", qualifier=absent).
- input = absent.

#### Linear chain — ancestors in root-first order

Setup:
- Create SPEC, SPEC/a, SPEC/a/b with empty frontmatter.

Actions:
1. Call ChainResolve("SPEC/a/b").

Expected:
- ancestors = [SPEC, SPEC/a] in that order.
- target = SPEC/a/b.

#### Single parent

Setup:
- Create SPEC, SPEC/a with empty frontmatter.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- ancestors = [SPEC].
- target = SPEC/a.

#### Target with empty frontmatter

Setup:
- Create SPEC, SPEC/a with empty frontmatter.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- ancestors = [SPEC], target = SPEC/a,
  dependencies = empty, input = absent.

### Dependencies — SPEC/ references

#### Dependency without qualifier

Setup:
- Create SPEC, SPEC/a (depends_on = ["SPEC/b"]),
  SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- dependencies contains one ChainItem with
  UnqualifiedLogicalName = "SPEC/b",
  Qualifier = absent.

#### Dependency with qualifier

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/b(interface)"]), SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- dependencies contains one ChainItem with
  UnqualifiedLogicalName = "SPEC/b",
  Qualifier = "interface".

#### Dependencies sorted by logical name

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/z", "SPEC/m", "SPEC/b"]),
  SPEC/z, SPEC/m, SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- dependencies = [SPEC/b, SPEC/m, SPEC/z] in that
  order.

### Dependencies — ARTIFACT/ references

#### ARTIFACT dependency resolved from generating node

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["ARTIFACT/b"]),
  SPEC/b (output = "out/lib.go").

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- dependencies contains one ChainItem with
  UnqualifiedLogicalName = "ARTIFACT/b",
  FilePath = "out/lib.go".

#### ARTIFACT — generating node has no output

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["ARTIFACT/b"]),
  SPEC/b (empty frontmatter, no output).

Actions:
1. Call ChainResolve("SPEC/a").

Expected: Error ErrUnresolvableArtifact.

#### ARTIFACT — artifact file does not exist on disk

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["ARTIFACT/b"]),
  SPEC/b (output = "out/lib.go").
- Do NOT create "out/lib.go" on disk.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- No error. dependencies contains one ChainItem
  with FilePath = "out/lib.go". Existence is not
  verified.

#### Mixed SPEC/, ARTIFACT/, EXTERNAL/ dependencies

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/c", "ARTIFACT/b",
  "EXTERNAL/proto/api.proto"]),
  SPEC/b (output = "out/lib.go"), SPEC/c.

Actions:
1. Call ChainResolve("SPEC/a").

Expected:
- dependencies sorted: ARTIFACT/b,
  EXTERNAL/proto/api.proto, SPEC/c.

### Dependencies — dedup

#### Exact duplicate — same file, same qualifier

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/b", "SPEC/b"]), SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains one entry for SPEC/b.

#### No qualifier subsumes qualifier

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/b", "SPEC/b(interface)"]),
  SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains one entry for SPEC/b
with Qualifier = absent.

#### Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/b(interface)", "SPEC/b"]),
  SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains one entry for SPEC/b
with Qualifier = absent.

#### Same file, different qualifiers — both kept

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["SPEC/b(interface)",
  "SPEC/b(constraints)"]), SPEC/b.

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains two entries, one
with Qualifier = "constraints", one with "interface"
(sorted).

#### Duplicate ARTIFACT — same logical name

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["ARTIFACT/b", "ARTIFACT/b"]),
  SPEC/b (output = "out/lib.go").

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains one ARTIFACT entry.

### Dependencies — EXTERNAL/ references

#### EXTERNAL dependency resolved to path

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["EXTERNAL/docs/api.yaml"]).

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains one ChainItem with
UnqualifiedLogicalName = "EXTERNAL/docs/api.yaml",
FilePath = "docs/api.yaml", Qualifier = absent.

#### Multiple EXTERNAL dependencies sorted

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["EXTERNAL/proto/v1.proto",
  "EXTERNAL/docs/api.yaml"]).

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies sorted:
EXTERNAL/docs/api.yaml, EXTERNAL/proto/v1.proto.

#### Duplicate EXTERNAL — same logical name

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["EXTERNAL/x.proto",
  "EXTERNAL/x.proto"]).

Actions:
1. Call ChainResolve("SPEC/a").

Expected: dependencies contains one EXTERNAL entry.

### Input

#### Input resolved from generating node

Setup:
- Create SPEC, SPEC/a (input = "ARTIFACT/b"),
  SPEC/b (output = "out/data.json").

Actions:
1. Call ChainResolve("SPEC/a").

Expected: input = ChainItem("ARTIFACT/b",
FilePath = "out/data.json").

#### EXTERNAL input resolved to path

Setup:
- Create SPEC, SPEC/a
  (input = "EXTERNAL/docs/vendor/spec.yaml").

Actions:
1. Call ChainResolve("SPEC/a").

Expected: input = ChainItem(
"EXTERNAL/docs/vendor/spec.yaml",
FilePath = "docs/vendor/spec.yaml").

#### No input — absent

Setup:
- Create SPEC, SPEC/a (no input).

Actions:
1. Call ChainResolve("SPEC/a").

Expected: input is absent (nil).

### Error cases

#### Unrecognized prefix in depends_on

Setup:
- Create SPEC, SPEC/a
  (depends_on = ["UNKNOWN/something"]).

Actions:
1. Call ChainResolve("SPEC/a").

Expected: Error ErrUnresolvableArtifact.

#### Invalid target logical name

Actions:
1. Call ChainResolve("INVALID/something").

Expected: Error propagated from LogicalNameParse.

#### Unreadable frontmatter

Setup:
- Create SPEC, SPEC/a with invalid YAML in frontmatter.

Actions:
1. Call ChainResolve("SPEC/a").

Expected: Error ErrUnreadableFrontmatter.

## Go-specific guidance

- The package name is `chainresolver_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
