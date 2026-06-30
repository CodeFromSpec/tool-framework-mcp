---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/test/utils/helpers
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/parsing(interface)
output: internal/chainhash/chainhash_test.go
---

# SPEC/golang/test/cases/chain/hash

# Agent

## Test setup guidance

`ChainHashCompute` calls `parsing.ParseNode` internally
for spec node positions (ancestors, target, SPEC/
dependencies). `parsing.ParseNode` requires a valid
`SPEC/` logical name that resolves to a `_node.md` file
on disk.

Therefore, tests that reference spec nodes must:
1. Use `testutils.Chdir` to set the working directory.
2. Create `code-from-spec/.../_node.md` files on disk
   matching the logical names used in CfsReferences.
3. Set `CfsReference.LogicalName` to a valid `SPEC/`
   logical name (e.g. `"SPEC/root/a"`), not a file path.
4. Set `CfsReference.Path` to the corresponding path
   (e.g. `"code-from-spec/root/a/_node.md"`).

For ARTIFACT/ items, `CfsReference.LogicalName` must
start with `"ARTIFACT/"` so the implementation reads
the file directly instead of calling
`parsing.ParseNode`.

Node files on disk must have valid structure for
`parsing.ParseNode`: at minimum a `# <logical_name>`
heading as the first heading.

## Test cases

### Properties

#### Hash is deterministic

Setup:
- Create a `_node.md` file for SPEC/root/a with `# Public`
  containing a `## Interface` subsection.
- Build a chainresolver.Chain with target = CfsReference pointing to
  SPEC/a.

Actions:
1. Call chainhash.ChainHashCompute twice with the same Chain.

Expected: Both chain hash results are identical strings.
Both positions slices have the same length and content.

#### Hash is 27 characters

Setup:
- Create a `_node.md` file for SPEC/root/a with `# Public`
  containing a `## Interface` subsection.
- Build a chainresolver.Chain with target = CfsReference pointing to
  that file.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: Chain hash is exactly 27 characters long.
Positions slice is non-empty.

#### Hash changes when ancestor content changes

Setup:
- Create `_node.md` for SPEC/root with `# Public` → `## Context`
  with initial content.
- Create `_node.md` for SPEC/root/a with `# Public` →
  `## Interface`.
- Build chainresolver.Chain with ancestors = [SPEC/root], target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/root's `## Context` content on disk.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Hash changes when dependency content changes

Setup:
- Create `_node.md` for SPEC/root with `# Public` → `## Context`.
- Create `_node.md` for SPEC/root/b with `# Public` →
  `## Interface` with initial content.
- Create `_node.md` for SPEC/a.
- Build chainresolver.Chain with target = SPEC/root/a, dependencies =
  [SPEC/root/b (qualifier absent)].

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/root/b's `## Interface` content on disk.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Hash changes when target Public changes

Setup:
- Create `_node.md` for SPEC/root with `# Public` → `## Context`.
- Create `_node.md` for SPEC/root/a with `# Public` →
  `## Interface` with initial content.
- Build chainresolver.Chain with target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/a's `## Interface` content on disk.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Hash changes when target Agent changes

Setup:
- Create `_node.md` for SPEC/root/a with `# Public` →
  `## Interface` and `# Agent` with initial content.
- Build chainresolver.Chain with target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/a's `# Agent` content on disk.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

### Ancestors

#### Ancestor with Public subsections contributes hash

Setup:
- Create SPEC/root with `# Public` → `## Context`.
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with ancestors = [SPEC/root], target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: Non-empty result of 27 characters.

#### Ancestor without Public section — skipped

Setup:
- Create SPEC/root/a with `# Public` → `## Interface`.
- Create SPEC/root with `# Public` → `## Context`.
- Build chainresolver.Chain with ancestors = [SPEC/root], target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute → hash_with_public.
2. Rewrite SPEC/root to have only a name heading (no
   `# Public`).
3. Call chainhash.ChainHashCompute → hash_without_public.

Expected: hash_with_public differs from
hash_without_public.

#### Multiple ancestors — order matters

Setup:
- Create SPEC/root with `# Public` → `## Context`
  ("root context").
- Create SPEC/root/a with `# Public` → `## Context`
  ("a context").
- Create SPEC/root/a/b with `# Public` → `## Interface`.
- Build Chain_A (chainresolver.Chain) with ancestors = [SPEC/root, SPEC/root/a],
  target = SPEC/root/a/b.
- Build Chain_B (chainresolver.Chain) with ancestors = [SPEC/root/a, SPEC/root],
  target = SPEC/root/a/b.

Actions:
1. Call chainhash.ChainHashCompute(Chain_A) → hash_a.
2. Call chainhash.ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a differs from hash_b.

### Dependencies

#### SPEC dependency without qualifier — hashes Public subsections

Setup:
- Create SPEC/root/b with `# Public` → `## Interface` with
  initial content.
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with target = SPEC/root/a, dependencies =
  [SPEC/root/b (qualifier absent)].

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/root/b's `## Interface` content.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### SPEC dependency with qualifier — hashes subsection

Setup:
- Create SPEC/root/b with `# Public` → `## Interface` with
  initial content.
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with target = SPEC/root/a, dependencies =
  [SPEC/root/b, qualifier = "interface"].

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/root/b's `## Interface` content.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Qualifier case normalization

Setup:
- Create SPEC/root/b with `## Interface` subsection.
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with dependency on SPEC/root/b,
  qualifier = "INTERFACE" (uppercase).

Actions:
1. Call chainhash.ChainHashCompute.

Expected: No error. Qualifier normalized before
matching.

#### ARTIFACT dependency — hashes full file content

Setup:
- Create an artifact file with content.
- Build chainresolver.Chain with ARTIFACT dependency pointing to
  that file.

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify the content.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### EXTERNAL dependency — hashes all content

Setup:
- Create an external file with initial content.
- Build chainresolver.Chain with EXTERNAL dependency pointing to
  that file.

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify the file.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

### Block extraction

#### Leading blank lines removed from subsection

Setup:
- Create file_A: `_node.md` with `## Interface` with
  two blank lines between heading and content.
- Create file_B: same content, no blank lines between
  heading and content.
- Build Chain_A and Chain_B (both chainresolver.Chain)
  targeting each.

Actions:
1. chainhash.ChainHashCompute(Chain_A) → hash_a.
2. chainhash.ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a equals hash_b.

#### Trailing blank lines removed from subsection

Setup:
- Create file_A: `_node.md` with `## Interface` with
  trailing blank lines.
- Create file_B: same content, no trailing blank lines.

Actions:
1. chainhash.ChainHashCompute(Chain_A) → hash_a.
2. chainhash.ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a equals hash_b.

#### Interior blank lines preserved

Setup:
- Create file_A: `_node.md` with `## Interface` with
  blank lines between content lines.
- Create file_B: same content, interior blank lines
  removed.

Actions:
1. chainhash.ChainHashCompute(Chain_A) → hash_a.
2. chainhash.ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a differs from hash_b.

### Target

#### Target Public and Agent both contribute

Setup:
- Create SPEC/root/a with `# Public` → `## Interface` and
  `# Agent` with content.
- Build chainresolver.Chain with target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Remove `# Agent` from file.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Target without Agent — Agent skipped

Setup:
- Create SPEC/root/a with `# Public` → `## Interface`, no
  `# Agent`.
- Build chainresolver.Chain with target = SPEC/root/a.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: No error. 27-character result.

### Input

#### Input hashes full file content

Setup:
- Create an artifact file with content.
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with target = SPEC/root/a, input =
  parsing.CfsReference(ARTIFACT/input, Path=<path>).

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify the artifact file.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### SPEC input — hashes Public subsections

Setup:
- Create SPEC/root/b with `# Public` → `## Interface`
  with initial content.
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with target = SPEC/root/a,
  input = parsing.CfsReference(SPEC/root/b,
  Qualifier=nil).

Actions:
1. Call chainhash.ChainHashCompute → hash_before.
2. Modify SPEC/root/b's `## Interface` content on disk.
3. Call chainhash.ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### No input — skipped

Setup:
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with target = SPEC/root/a, input absent.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: No error. 27-character result.

### Qualifier edge cases

#### Qualifier references non-existent subsection

Setup:
- Create SPEC/root/b with `# Public` → `## Context`
  only (no `## Interface`).
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with target = SPEC/root/a,
  dependencies = [SPEC/root/b, qualifier = "interface"].

Actions:
1. Call chainhash.ChainHashCompute.

Expected: No error. The non-matching qualifier
position is silently skipped.

### Error cases

#### Unreadable spec node file

Setup:
- Build chainresolver.Chain with target pointing to a non-existent
  file.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: chainhash.ErrParseFailure.

#### Unreadable artifact file

Setup:
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with ARTIFACT dependency pointing to a
  non-existent file.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: oslayer.ErrFileUnreadable.

#### Unreadable external file

Setup:
- Create SPEC/root/a with `# Public` → `## Interface`.
- Build chainresolver.Chain with EXTERNAL dependency pointing to a
  non-existent file.

Actions:
1. Call chainhash.ChainHashCompute.

Expected: oslayer.ErrFileUnreadable.

## Go-specific guidance

- The package name is `chainhash_test` (external test
  package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
