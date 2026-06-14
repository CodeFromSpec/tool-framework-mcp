---
depends_on:
  - SPEC/functional/logic/chain/resolver(interface)
output: code-from-spec/functional/tests/chain/resolver/output.md
---

# SPEC/functional/tests/chain/resolver

Test cases for the chain resolver component.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files containing frontmatter as needed, then call
`ChainResolve` with a target logical name.

### Ancestors and target

#### Root as target

Create spec tree: SPEC only. Call ChainResolve with
"SPEC". Expect ancestors = empty, target =
ChainItem(unqualified_logical_name="SPEC", qualifier=absent).

#### Linear chain — ancestors in root-first order

Create spec tree: SPEC, SPEC/a, SPEC/a/b. Call
ChainResolve with "SPEC/a/b". Expect ancestors =
[SPEC, SPEC/a] in that order. Target = SPEC/a/b.

#### Single parent

Create spec tree: SPEC, SPEC/a. Call ChainResolve with
"SPEC/a". Expect ancestors = [SPEC]. Target = SPEC/a.

#### Target with empty frontmatter

Create spec tree: SPEC, SPEC/a (leaf, empty
frontmatter). Call ChainResolve with "SPEC/a".

Expect ancestors = [SPEC], target = SPEC/a,
dependencies = empty, input = absent.

### Dependencies — SPEC/ references

#### Dependency without qualifier

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/b"]), SPEC/b. Call ChainResolve with "SPEC/a".

Expect dependencies contains one ChainItem with
unqualified_logical_name = "SPEC/b", qualifier = absent.

#### Dependency with qualifier

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/b(interface)"]), SPEC/b. Call ChainResolve with
"SPEC/a".

Expect dependencies contains one ChainItem with
unqualified_logical_name = "SPEC/b", qualifier = "interface".

#### Dependencies sorted by logical name then qualifier

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/z", "SPEC/m", "SPEC/b"]). Create SPEC/z, SPEC/m,
SPEC/b. Call ChainResolve with "SPEC/a".

Expect dependencies sorted alphabetically by logical
name: SPEC/b, SPEC/m, SPEC/z.

### Dependencies — ARTIFACT/ references

#### ARTIFACT dependency resolved from generating node

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["ARTIFACT/b"]), SPEC/b (with output = "out/lib.go").
Call ChainResolve with "SPEC/a".

Expect dependencies contains one ChainItem with
unqualified_logical_name = "ARTIFACT/b", file_path =
"out/lib.go".

#### ARTIFACT — generating node has no output

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["ARTIFACT/b"]), SPEC/b (with empty frontmatter,
no output). Call ChainResolve with "SPEC/a".

Expect error UnresolvableArtifact.

#### ARTIFACT — artifact file does not exist on disk

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["ARTIFACT/b"]), SPEC/b (with output = "out/lib.go").
Do NOT create "out/lib.go" on disk. Call ChainResolve
with "SPEC/a".

Expect no error. Dependencies contains one ChainItem
with file_path = "out/lib.go". Existence is not
verified by the resolver.

#### Mixed SPEC/, ARTIFACT/, and EXTERNAL/ dependencies

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/c", "ARTIFACT/b", "EXTERNAL/proto/api.proto"]),
SPEC/b (with output = "out/lib.go"), SPEC/c. Call
ChainResolve with "SPEC/a".

Expect dependencies contains all three entries, sorted
by logical name: ARTIFACT/b, EXTERNAL/proto/api.proto,
SPEC/c.

### Dependencies — dedup

#### Exact duplicate — same file, same qualifier

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/b", "SPEC/b"]), SPEC/b. Call ChainResolve with
"SPEC/a".

Expect dependencies contains one entry for SPEC/b
(not two).

#### No qualifier subsumes qualifier

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/b", "SPEC/b(interface)"]), SPEC/b. Call
ChainResolve with "SPEC/a".

Expect dependencies contains one entry for SPEC/b
with qualifier = absent. The SPEC/b(interface) entry
is removed.

#### Qualifier before no-qualifier — no-qualifier wins

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/b(interface)", "SPEC/b"]), SPEC/b. Call
ChainResolve with "SPEC/a".

Expect dependencies contains one entry for SPEC/b
with qualifier = absent.

#### Same file, different qualifiers — both kept

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["SPEC/b(interface)", "SPEC/b(constraints)"]), SPEC/b.
Call ChainResolve with "SPEC/a".

Expect dependencies contains two entries, one with
qualifier = "constraints", one with "interface"
(sorted).

#### Duplicate ARTIFACT — same logical name

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["ARTIFACT/b", "ARTIFACT/b"]), SPEC/b (with output =
"out/lib.go"). Call ChainResolve with "SPEC/a".

Expect dependencies contains one ARTIFACT entry
(not two).

### Dependencies — EXTERNAL/ references

#### EXTERNAL dependency resolved to path

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["EXTERNAL/docs/api.yaml"]). Call ChainResolve with
"SPEC/a".

Expect dependencies contains one ChainItem with
unqualified_logical_name = "EXTERNAL/docs/api.yaml", file_path =
"docs/api.yaml", qualifier = absent.

#### Multiple EXTERNAL dependencies sorted

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["EXTERNAL/proto/v1.proto",
"EXTERNAL/docs/api.yaml"]). Call ChainResolve with
"SPEC/a".

Expect dependencies sorted by logical name:
EXTERNAL/docs/api.yaml before
EXTERNAL/proto/v1.proto.

#### Duplicate EXTERNAL — same logical name

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["EXTERNAL/x.proto", "EXTERNAL/x.proto"]). Call
ChainResolve with "SPEC/a".

Expect dependencies contains one EXTERNAL entry
(not two).

### Input

#### Input resolved from generating node

Create spec tree: SPEC, SPEC/a (leaf, input =
"ARTIFACT/b"), SPEC/b (with output =
"out/data.json"). Call ChainResolve with "SPEC/a".

Expect input = ChainItem with unqualified_logical_name =
"ARTIFACT/b", file_path = "out/data.json".

#### EXTERNAL input resolved to path

Create spec tree: SPEC, SPEC/a (leaf, input =
"EXTERNAL/docs/vendor/spec.yaml"). Call ChainResolve
with "SPEC/a".

Expect input = ChainItem with unqualified_logical_name =
"EXTERNAL/docs/vendor/spec.yaml", file_path =
"docs/vendor/spec.yaml".

#### No input — absent

Create spec tree: SPEC, SPEC/a (leaf, no input). Call
ChainResolve with "SPEC/a".

Expect input is absent.

### Error cases

#### Unrecognized prefix in depends_on

Create spec tree: SPEC, SPEC/a (leaf, depends_on =
["UNKNOWN/something"]). Call ChainResolve with "SPEC/a".

Expect error UnresolvableArtifact.

#### Invalid target logical name

Call ChainResolve with "INVALID/something".

Expect error propagated from LogicalNameGetParent or
LogicalNameToPath.

#### Unreadable frontmatter

Create spec tree: SPEC, SPEC/a with invalid YAML in
frontmatter. Call ChainResolve with "SPEC/a".

Expect error UnreadableFrontmatter.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `ChainResolve`.
- Use the record names from the interface: `ChainItem`,
  `Chain`.
- Each test case creates a spec tree on disk with
  `_node.md` files, then calls `ChainResolve`.
- Describe setup as files to create with their
  frontmatter content.
