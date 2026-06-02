---
depends_on:
  - ROOT/functional/logic/chain/hash(interface)
  - ROOT/functional/logic/chain/resolver(interface)
output: code-from-spec/functional/tests/chain/hash/output.md
---

# ROOT/functional/tests/chain/hash

Test cases for the chain hash component.

All tests build a `Chain` record (as returned by
`ChainResolve`) and create the referenced files on
disk, then call `ChainHashCompute`.

# Public

## Test cases

### Properties

#### Hash is deterministic

Create files on disk. Build a Chain. Call
ChainHashCompute twice with the same Chain. Expect
both results are identical.

#### Hash is 27 characters

Call ChainHashCompute with any valid Chain. Expect the
result is exactly 27 characters long.

#### Hash changes when ancestor content changes

Create a spec tree: ROOT with `# Public` content,
ROOT/a as target. Build a Chain with ROOT as ancestor.
Compute hash. Modify ROOT's `# Public` content on
disk. Recompute. Expect hashes differ.

#### Hash changes when dependency content changes

Create a spec tree: ROOT, ROOT/a, ROOT/b. ROOT/a is
target with dependency on ROOT/b. Build Chain. Compute
hash. Modify ROOT/b's `# Public` content. Recompute.
Expect hashes differ.

#### Hash changes when target Public changes

Create a spec tree: ROOT, ROOT/a as target with
`# Public` content. Build Chain. Compute hash. Modify
ROOT/a's `# Public` on disk. Recompute. Expect hashes
differ.

#### Hash changes when target Agent changes

Create a spec tree: ROOT, ROOT/a as target with
`# Agent` content. Build Chain. Compute hash. Modify
ROOT/a's `# Agent` on disk. Recompute. Expect hashes
differ.

### Ancestors

#### Ancestor with Public section contributes hash

Create ROOT with `# Public` content, ROOT/a as target.
Build Chain with ancestors = [ROOT]. Compute hash.
Expect a non-empty result (27 chars).

#### Ancestor without Public section — skipped

Create ROOT with no `# Public` section (only name
section). Build Chain with ancestors = [ROOT]. Compute
hash. The result should differ from a chain with an
ancestor that has `# Public`.

#### Multiple ancestors — order matters

Create ROOT, ROOT/a, ROOT/a/b as target. ROOT and
ROOT/a both have `# Public`. Build Chain with
ancestors = [ROOT, ROOT/a] in root-first order.
Compute hash. Swap ancestor order and recompute.
Expect hashes differ.

### Dependencies

#### ROOT dependency without qualifier — hashes Public

Create ROOT/b with `# Public` content. Build Chain
with dependency on ROOT/b (qualifier absent). Compute
hash. Modify ROOT/b's `# Public`. Recompute. Expect
hashes differ.

#### ROOT dependency with qualifier — hashes subsection

Create ROOT/b with `# Public` containing `## Interface`
subsection. Build Chain with dependency on ROOT/b,
qualifier = "interface". Compute hash. Modify the
`## Interface` content. Recompute. Expect hashes differ.

#### Qualifier case normalization

Create ROOT/b with `## Interface` subsection. Build
Chain with qualifier = "INTERFACE" (uppercase). Compute
hash. Expect no error — the qualifier is normalized
before matching.

#### ARTIFACT dependency — hashes file minus frontmatter

Create an artifact file with frontmatter and body
content. Build Chain with ARTIFACT dependency pointing
to that file. Compute hash. Modify the body. Recompute.
Expect hashes differ.

#### ARTIFACT dependency — frontmatter change ignored

Create an artifact file with frontmatter and body.
Compute hash. Modify only the frontmatter. Recompute.
Expect hashes are identical — frontmatter is stripped.

### External files

#### External file — hashes all content

Create an external file. Build Chain with external
entry. Compute hash. Modify the file. Recompute.
Expect hashes differ.

### Target

#### Target Public and Agent both contribute

Create ROOT/a as target with `# Public` and `# Agent`.
Build Chain. Compute hash. Remove `# Agent` from file.
Recompute. Expect hashes differ.

#### Target without Agent — Agent skipped

Create ROOT/a as target with `# Public` only, no
`# Agent`. Build Chain. Compute hash. Expect no error.

### Input

#### Input hashes file minus frontmatter

Create an artifact file with frontmatter and body.
Build Chain with input pointing to that file. Compute
hash. Modify the body. Recompute. Expect hashes differ.

#### No input — skipped

Build Chain with input absent. Compute hash. Expect
no error.

### Error cases

#### Unreadable spec node file

Build Chain referencing a spec node whose file does not
exist on disk. Call ChainHashCompute. Expect error
ParseFailure.

#### Unreadable artifact file

Build Chain with ARTIFACT dependency pointing to a
non-existent file. Call ChainHashCompute. Expect error
FileUnreadable.

#### Unreadable external file

Build Chain with external entry pointing to a
non-existent file. Call ChainHashCompute. Expect error
FileUnreadable.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `ChainHashCompute`.
- Use the record names from the interface: `Chain`,
  `ChainItem`.
- Tests build `Chain` records directly — they do not
  call `ChainResolve`.
- Each test creates files on disk as needed.
