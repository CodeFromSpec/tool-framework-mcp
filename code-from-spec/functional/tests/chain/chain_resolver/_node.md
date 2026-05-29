---
depends_on:
  - ROOT/functional/logic/chain/chain_resolver(interface)
outputs:
  - id: chain_resolver_tests
    path: code-from-spec/functional/tests/chain/chain_resolver/output.md
---

# ROOT/functional/tests/chain/chain_resolver

Test cases for the chain resolver component.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files containing frontmatter as needed, then call
`ChainResolve` with a target logical name.

### Ancestors and target

#### Root as target

Create spec tree: ROOT only. Call ChainResolve with
"ROOT". Expect ancestors = empty, target =
ChainItem(logical_name="ROOT", qualifier=absent).

#### Linear chain — ancestors in root-first order

Create spec tree: ROOT, ROOT/a, ROOT/a/b. Call
ChainResolve with "ROOT/a/b". Expect ancestors =
[ROOT, ROOT/a] in that order. Target = ROOT/a/b.

#### Single parent

Create spec tree: ROOT, ROOT/a. Call ChainResolve with
"ROOT/a". Expect ancestors = [ROOT]. Target = ROOT/a.

#### Target with empty frontmatter

Create spec tree: ROOT, ROOT/a (leaf, empty
frontmatter). Call ChainResolve with "ROOT/a".

Expect ancestors = [ROOT], target = ROOT/a,
dependencies = empty, external = empty, input = absent.

### Dependencies — ROOT/ references

#### Dependency without qualifier

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/b"]), ROOT/b. Call ChainResolve with "ROOT/a".

Expect dependencies contains one ChainItem with
logical_name = "ROOT/b", qualifier = absent.

#### Dependency with qualifier

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/b(interface)"]), ROOT/b. Call ChainResolve with
"ROOT/a".

Expect dependencies contains one ChainItem with
logical_name = "ROOT/b", qualifier = "interface".

#### Dependencies sorted by file path then qualifier

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/z", "ROOT/m", "ROOT/b"]). Create ROOT/z, ROOT/m,
ROOT/b. Call ChainResolve with "ROOT/a".

Expect dependencies sorted alphabetically by file path.

### Dependencies — ARTIFACT/ references

#### ARTIFACT dependency resolved from generating node

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/b(lib)"]), ROOT/b (with outputs = [{id:
"lib", path: "out/lib.go"}]). Call ChainResolve with
"ROOT/a".

Expect dependencies contains one ChainItem with
logical_name = "ARTIFACT/b(lib)", file_path =
"out/lib.go", qualifier = "lib".

#### ARTIFACT without qualifier — error

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/b"]). Call ChainResolve with "ROOT/a".

Expect error "unresolvable artifact".

#### ARTIFACT — generating node has no outputs

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/b(lib)"]), ROOT/b (with empty frontmatter,
no outputs). Call ChainResolve with "ROOT/a".

Expect error "unresolvable artifact".

#### ARTIFACT — artifact file does not exist on disk

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/b(lib)"]), ROOT/b (with outputs = [{id:
"lib", path: "out/lib.go"}]). Do NOT create
"out/lib.go" on disk. Call ChainResolve with "ROOT/a".

Expect no error. Dependencies contains one ChainItem
with file_path = "out/lib.go". Existence is not
verified by the resolver.

#### ARTIFACT with non-existent output id — error

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/b(missing)"]), ROOT/b (with outputs =
[{id: "lib", path: "out/lib.go"}]). Call ChainResolve
with "ROOT/a".

Expect error "unresolvable artifact".

#### Mixed ROOT/ and ARTIFACT/ dependencies

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/c", "ARTIFACT/b(lib)"]), ROOT/b (with outputs =
[{id: "lib", path: "out/lib.go"}]), ROOT/c. Call
ChainResolve with "ROOT/a".

Expect dependencies contains both entries, sorted by
file path value.

### Dependencies — dedup

#### Exact duplicate — same file, same qualifier

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/b", "ROOT/b"]), ROOT/b. Call ChainResolve with
"ROOT/a".

Expect dependencies contains one entry for ROOT/b
(not two).

#### No qualifier subsumes qualifier

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/b", "ROOT/b(interface)"]), ROOT/b. Call
ChainResolve with "ROOT/a".

Expect dependencies contains one entry for ROOT/b
with qualifier = absent. The ROOT/b(interface) entry
is removed.

#### Qualifier before no-qualifier — no-qualifier wins

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/b(interface)", "ROOT/b"]), ROOT/b. Call
ChainResolve with "ROOT/a".

Expect dependencies contains one entry for ROOT/b
with qualifier = absent.

#### Same file, different qualifiers — both kept

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ROOT/b(interface)", "ROOT/b(constraints)"]), ROOT/b.
Call ChainResolve with "ROOT/a".

Expect dependencies contains two entries, one with
qualifier = "constraints", one with "interface"
(sorted).

#### Duplicate ARTIFACT — same logical name

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/b(lib)", "ARTIFACT/b(lib)"]), ROOT/b (with
outputs = [{id: "lib", path: "out/lib.go"}]). Call
ChainResolve with "ROOT/a".

Expect dependencies contains one ARTIFACT entry
(not two).

### External

#### External entries copied from frontmatter

Create spec tree: ROOT, ROOT/a (leaf, external =
[{path: "docs/api.yaml"}, {path: "proto/v1.proto"}]).
Call ChainResolve with "ROOT/a".

Expect external list contains both entries, sorted
by path: proto/v1.proto before docs/api.yaml? No —
alphabetically: docs/api.yaml before proto/v1.proto.

#### External with fragments preserved

Create spec tree: ROOT, ROOT/a (leaf, external =
[{path: "f.txt", fragments: [{lines: "1-10",
hash: "abc"}]}]). Call ChainResolve with "ROOT/a".

Expect external list contains one entry with path =
"f.txt" and fragments preserved as-is.

#### Empty external — no entries

Create spec tree: ROOT, ROOT/a (leaf, no external).
Call ChainResolve with "ROOT/a".

Expect external list is empty.

### Input

#### Input resolved from generating node

Create spec tree: ROOT, ROOT/a (leaf, input =
"ARTIFACT/b(data)"), ROOT/b (with outputs = [{id:
"data", path: "out/data.json"}]). Call ChainResolve
with "ROOT/a".

Expect input = ChainItem with logical_name =
"ARTIFACT/b(data)", file_path = "out/data.json",
qualifier = "data".

#### No input — absent

Create spec tree: ROOT, ROOT/a (leaf, no input). Call
ChainResolve with "ROOT/a".

Expect input is absent.

#### Input without qualifier — error

Create spec tree: ROOT, ROOT/a (leaf, input =
"ARTIFACT/b"). Call ChainResolve with "ROOT/a".

Expect error "unresolvable artifact".

#### Input with non-existent output id — error

Create spec tree: ROOT, ROOT/a (leaf, input =
"ARTIFACT/b(missing)"), ROOT/b (with outputs = [{id:
"data", path: "out/data.json"}]). Call ChainResolve
with "ROOT/a".

Expect error "unresolvable artifact".

### Error cases

#### Unrecognized prefix in depends_on

Create spec tree: ROOT, ROOT/a (leaf, depends_on =
["UNKNOWN/something"]). Call ChainResolve with "ROOT/a".

Expect error "unresolvable artifact".

#### Invalid target logical name

Call ChainResolve with "INVALID/something".

Expect error propagated from LogicalNameGetParent or
LogicalNameToPath.

#### Unreadable frontmatter

Create spec tree: ROOT, ROOT/a with invalid YAML in
frontmatter. Call ChainResolve with "ROOT/a".

Expect error "unreadable frontmatter".

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
