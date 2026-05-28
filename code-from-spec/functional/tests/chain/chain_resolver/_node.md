---
depends_on:
  - ROOT/functional/logic/chain/chain_resolver(interface)
outputs:
  - id: chain_resolver_tests
    path: code-from-spec/functional/tests/chain/chain_resolver/output.md
---

# ROOT/functional/tests/chain/chain_resolver

Test cases for the chain resolver component.

Review status: pending

# Public

## Test cases

### Happy path

#### Leaf node -- ancestors only, no dependencies

Create a spec tree: ROOT, ROOT/a, ROOT/a/b (leaf).
Call ResolveChain with "ROOT/a/b".

Expect ancestors = ROOT, ROOT/a (sorted alphabetically),
each with no qualifier. Target = ROOT/a/b with no
qualifier. Dependencies empty. Code empty.

#### Leaf node -- with dependency, no qualifier

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b), ROOT/b. Call ResolveChain with "ROOT/a".

Expect ancestors = ROOT. Target = ROOT/a. Dependencies
contains one item ROOT/b with no qualifier.

#### Leaf node -- with dependency, with qualifier

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b(interface)), ROOT/b. Call ResolveChain with
"ROOT/a".

Expect dependencies contains one item with logical name =
"ROOT/b(interface)", file path pointing to ROOT/b's node
file, qualifier = "interface".

#### Dependencies sorted

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/z, ROOT/m, ROOT/b). Call ResolveChain with "ROOT/a".

Expect dependencies sorted by file path.

#### Leaf node -- outputs file exists on disk

Create a spec tree: ROOT, ROOT/a (leaf with outputs
pointing to src/a.go). Create the file src/a.go on disk.
Call ResolveChain with "ROOT/a".

Expect code = ["src/a.go"].

#### Leaf node -- outputs file does not exist

Create a spec tree: ROOT, ROOT/a (leaf with outputs
pointing to src/a.go). Do not create src/a.go.
Call ResolveChain with "ROOT/a".

Expect code is empty.

#### Multiple qualifiers for same file

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b(interface) and ROOT/b(constraints)), ROOT/b.
Call ResolveChain with "ROOT/a".

Expect dependencies contains two items, both pointing to
ROOT/b's file, one with qualifier = "interface", the other
with qualifier = "constraints".

### Edge cases

#### Dedup: same file, same qualifier

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b, ROOT/b), ROOT/b. Call ResolveChain with "ROOT/a".

Expect dependencies contains one item ROOT/b with no
qualifier (not two).

#### Dedup: same file, different qualifiers -- both kept

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b(interface), ROOT/b(constraints)), ROOT/b.
Call ResolveChain with "ROOT/a".

Expect dependencies contains two items for ROOT/b, one
with qualifier = "interface", one with "constraints".

#### Dedup: nil qualifier subsumes specific qualifiers

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b, ROOT/b(interface)), ROOT/b. Call ResolveChain
with "ROOT/a".

Expect dependencies contains one item ROOT/b with no
qualifier. The ROOT/b(interface) entry is removed because
no qualifier (whole public section) already includes the
specific qualifier.

#### Dedup: specific qualifier appears before nil -- nil wins

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b(interface), ROOT/b), ROOT/b. Call ResolveChain
with "ROOT/a".

Expect dependencies contains one item ROOT/b with no
qualifier. Even though the specific qualifier appeared
first, the nil entry subsumes it.

#### Dedup: repeated qualifier for same file

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/b(interface), ROOT/b(interface)), ROOT/b.
Call ResolveChain with "ROOT/a".

Expect dependencies contains one item with qualifier =
"interface" (not two).

### Failure cases

#### Invalid logical name

Call ResolveChain with "INVALID/something".

Expect error containing "cannot resolve logical name".

#### Unreadable frontmatter

Create a spec tree: ROOT, ROOT/a (leaf). Write invalid
YAML in ROOT/a's frontmatter. Call ResolveChain with
"ROOT/a".

Expect error from frontmatter parsing.

#### Unresolvable dependency

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
ROOT/nonexistent). Call ResolveChain with "ROOT/a".

Expect error containing "cannot resolve logical name".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `ResolveChain`.
