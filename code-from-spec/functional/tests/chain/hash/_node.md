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

Create a spec tree: SPEC with `# Public` containing a
`## Context` subsection, SPEC/a as target. Build a
Chain with SPEC as ancestor. Compute hash. Modify
SPEC's `## Context` subsection content on disk.
Recompute. Expect hashes differ.

#### Hash changes when dependency content changes

Create a spec tree: SPEC, SPEC/a, SPEC/b with
`# Public` containing a `## Interface` subsection.
SPEC/a is target with dependency on SPEC/b. Build
Chain. Compute hash. Modify SPEC/b's `## Interface`
subsection content. Recompute. Expect hashes differ.

#### Hash changes when target Public changes

Create a spec tree: SPEC, SPEC/a as target with
`# Public` containing a `## Interface` subsection.
Build Chain. Compute hash. Modify SPEC/a's
`## Interface` subsection content on disk. Recompute.
Expect hashes differ.

#### Hash changes when target Agent changes

Create a spec tree: SPEC, SPEC/a as target with
`# Agent` content. Build Chain. Compute hash. Modify
SPEC/a's `# Agent` on disk. Recompute. Expect hashes
differ.

### Ancestors

#### Ancestor with Public subsections contributes hash

Create SPEC with `# Public` containing a `## Context`
subsection, SPEC/a as target. Build Chain with
ancestors = [SPEC]. Compute hash. Expect a non-empty
result (27 chars).

#### Ancestor without Public section — skipped

Create SPEC/a as target with `# Public` containing a
`## Interface` subsection. First, create SPEC with
`# Public` containing a `## Context` subsection. Build
Chain with ancestors = [SPEC], target = SPEC/a. Compute
hash → hash_with_public. Then rewrite SPEC on disk to
have only a name section (no `# Public`). Build the
same Chain structure. Compute hash → hash_without_public.
Expect hash_with_public differs from hash_without_public.

#### Multiple ancestors — order matters

Create SPEC, SPEC/a, SPEC/a/b as target. SPEC and
SPEC/a both have `# Public` with `## Context`
subsections. Build Chain with ancestors = [SPEC,
SPEC/a] in root-first order. Compute hash. Swap
ancestor order and recompute. Expect hashes differ.

### Dependencies

#### SPEC dependency without qualifier — hashes Public subsections

Create SPEC/b with `# Public` containing a
`## Interface` subsection. Build Chain with dependency
on SPEC/b (qualifier absent). Compute hash. Modify
SPEC/b's `## Interface` subsection content. Recompute.
Expect hashes differ.

#### SPEC dependency with qualifier — hashes subsection

Create SPEC/b with `# Public` containing `## Interface`
subsection. Build Chain with dependency on SPEC/b,
qualifier = "interface". Compute hash. Modify the
`## Interface` content. Recompute. Expect hashes differ.

#### Qualifier case normalization

Create SPEC/b with `## Interface` subsection. Build
Chain with qualifier = "INTERFACE" (uppercase). Compute
hash. Expect no error — the qualifier is normalized
before matching.

#### ARTIFACT dependency — hashes full file content

Create an artifact file with content. Build Chain with
ARTIFACT dependency pointing to that file. Compute hash.
Modify the content. Recompute. Expect hashes differ.

#### ARTIFACT dependency — tag hash change ignored

Create an artifact file with body containing an artifact
tag line `// code-from-spec: SPEC/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn`.
Build Chain with ARTIFACT dependency pointing to that
file. Compute hash. Change only the 27-character hash
in the tag to a different value (e.g.
`zZyYxXwWvVuUtTsSrRqQpPoOnNm`). Recompute. Expect
hashes are identical — the tag hash is neutralized
before hashing.

#### EXTERNAL dependency — hashes all content

Create an external file. Build Chain with EXTERNAL
dependency pointing to that file. Compute hash. Modify
the file. Recompute. Expect hashes differ.

### Block extraction

#### Leading blank lines removed from subsection

Create a `_node.md` with a `## Interface` subsection
that has two blank lines between the heading and the
first content line. Build Chain referencing this node.
Compute hash. Create a second version with no blank
lines between heading and content (same content).
Compute hash. Expect hashes are identical — leading
blank lines are removed by block extraction.

#### Trailing blank lines removed from subsection

Create a `_node.md` with a `## Interface` subsection
that has trailing blank lines after the last content
line. Build Chain referencing this node. Compute hash.
Create a second version without trailing blank lines.
Compute hash. Expect hashes are identical.

#### Interior blank lines preserved

Create a `_node.md` with a `## Interface` subsection
that has blank lines between content lines. Build
Chain referencing this node. Compute hash. Remove the
interior blank lines. Compute hash. Expect hashes
differ — interior content is preserved byte for byte.

### Target

#### Target Public and Agent both contribute

Create SPEC/a as target with `# Public` containing a
`## Interface` subsection and `# Agent` with content.
Build Chain. Compute hash. Remove `# Agent` from file.
Recompute. Expect hashes differ.

#### Target without Agent — Agent skipped

Create SPEC/a as target with `# Public` containing a
`## Interface` subsection, no `# Agent`. Build Chain.
Compute hash. Expect no error.

### Input

#### Input hashes full file content

Create an artifact file with content. Build Chain with
input = ChainItem(unqualified_logical_name="ARTIFACT/input",
file_path=<path to the artifact file>). Compute hash.
Modify the file content. Recompute. Expect hashes
differ.

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
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
  Never place content directly under `# Public`
  without a subsection heading — this is a format
  error.
