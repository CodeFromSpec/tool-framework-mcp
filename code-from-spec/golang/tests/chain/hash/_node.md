---
depends_on:
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/utils/logical_names
output: internal/chainhash/chainhash_test.go
---

# SPEC/golang/tests/chain/hash

# Agent

## Test setup guidance

`ChainHashCompute` calls `NodeParse` internally for spec
node positions (ancestors, target, ROOT/ dependencies).
`NodeParse` requires a valid `ROOT/` logical name that
resolves to a `_node.md` file on disk.

Therefore, tests that reference spec nodes must:
1. Use `testChdir` to set the working directory.
2. Create `code-from-spec/.../_node.md` files on disk
   matching the logical names used in ChainItems.
3. Set `ChainItem.LogicalName` to a valid `ROOT/`
   logical name (e.g. `"ROOT/a"`), not a file path.
4. Set `ChainItem.FilePath` to the corresponding
   `PathCfs` (e.g. `{Value: "code-from-spec/a/_node.md"}`).

For ARTIFACT/ items, `ChainItem.LogicalName` must start
with `"ARTIFACT/"` so the implementation reads the file
directly instead of calling `NodeParse`.

Node files on disk must have valid structure for
`NodeParse`: at minimum a `# <logical_name>` heading
as the first heading.

## Test cases

### Properties

#### Hash is deterministic

Setup:
- Create a `_node.md` file for SPEC/a with `# Public`
  containing a `## Interface` subsection.
- Build a Chain with target = ChainItem pointing to
  SPEC/a.

Actions:
1. Call ChainHashCompute twice with the same Chain.

Expected: Both results are identical strings.

#### Hash is 27 characters

Setup:
- Create a `_node.md` file for SPEC/a with `# Public`
  containing a `## Interface` subsection.
- Build a Chain with target = ChainItem pointing to
  that file.

Actions:
1. Call ChainHashCompute.

Expected: Result is exactly 27 characters long.

#### Hash changes when ancestor content changes

Setup:
- Create `_node.md` for SPEC with `# Public` → `## Context`
  with initial content.
- Create `_node.md` for SPEC/a with `# Public` →
  `## Interface`.
- Build Chain with ancestors = [SPEC], target = SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC's `## Context` content on disk.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Hash changes when dependency content changes

Setup:
- Create `_node.md` for SPEC with `# Public` → `## Context`.
- Create `_node.md` for SPEC/b with `# Public` →
  `## Interface` with initial content.
- Create `_node.md` for SPEC/a.
- Build Chain with target = SPEC/a, dependencies =
  [SPEC/b (qualifier absent)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/b's `## Interface` content on disk.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Hash changes when target Public changes

Setup:
- Create `_node.md` for SPEC with `# Public` → `## Context`.
- Create `_node.md` for SPEC/a with `# Public` →
  `## Interface` with initial content.
- Build Chain with target = SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/a's `## Interface` content on disk.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Hash changes when target Agent changes

Setup:
- Create `_node.md` for SPEC/a with `# Public` →
  `## Interface` and `# Agent` with initial content.
- Build Chain with target = SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/a's `# Agent` content on disk.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

### Ancestors

#### Ancestor with Public subsections contributes hash

Setup:
- Create SPEC with `# Public` → `## Context`.
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with ancestors = [SPEC], target = SPEC/a.

Actions:
1. Call ChainHashCompute.

Expected: Non-empty result of 27 characters.

#### Ancestor without Public section — skipped

Setup:
- Create SPEC/a with `# Public` → `## Interface`.
- Create SPEC with `# Public` → `## Context`.
- Build Chain with ancestors = [SPEC], target = SPEC/a.

Actions:
1. Call ChainHashCompute → hash_with_public.
2. Rewrite SPEC to have only a name heading (no
   `# Public`).
3. Call ChainHashCompute → hash_without_public.

Expected: hash_with_public differs from
hash_without_public.

#### Multiple ancestors — order matters

Setup:
- Create SPEC with `# Public` → `## Context`
  ("root context").
- Create SPEC/a with `# Public` → `## Context`
  ("a context").
- Create SPEC/a/b with `# Public` → `## Interface`.
- Build Chain_A with ancestors = [SPEC, SPEC/a],
  target = SPEC/a/b.
- Build Chain_B with ancestors = [SPEC/a, SPEC],
  target = SPEC/a/b.

Actions:
1. Call ChainHashCompute(Chain_A) → hash_a.
2. Call ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a differs from hash_b.

### Dependencies

#### SPEC dependency without qualifier — hashes Public subsections

Setup:
- Create SPEC/b with `# Public` → `## Interface` with
  initial content.
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with target = SPEC/a, dependencies =
  [SPEC/b (qualifier absent)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/b's `## Interface` content.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### SPEC dependency with qualifier — hashes subsection

Setup:
- Create SPEC/b with `# Public` → `## Interface` with
  initial content.
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with target = SPEC/a, dependencies =
  [SPEC/b, qualifier = "interface"].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/b's `## Interface` content.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Qualifier case normalization

Setup:
- Create SPEC/b with `## Interface` subsection.
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with dependency on SPEC/b,
  qualifier = "INTERFACE" (uppercase).

Actions:
1. Call ChainHashCompute.

Expected: No error. Qualifier normalized before
matching.

#### ARTIFACT dependency — hashes full file content

Setup:
- Create an artifact file with content.
- Build Chain with ARTIFACT dependency pointing to
  that file.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the content.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### ARTIFACT dependency — tag hash change ignored

Setup:
- Create an artifact file containing:
  `// code-from-spec: SPEC/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn`
- Build Chain with ARTIFACT dependency pointing to
  that file.

Actions:
1. Call ChainHashCompute → hash_before.
2. Change only the 27-character hash in the tag to a
   different value. No other content change.
3. Call ChainHashCompute → hash_after.

Expected: hash_before equals hash_after — the tag
hash is neutralized.

#### EXTERNAL dependency — hashes all content

Setup:
- Create an external file with initial content.
- Build Chain with EXTERNAL dependency pointing to
  that file.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the file.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

### Block extraction

#### Leading blank lines removed from subsection

Setup:
- Create file_A: `_node.md` with `## Interface` with
  two blank lines between heading and content.
- Create file_B: same content, no blank lines between
  heading and content.
- Build Chain_A and Chain_B targeting each.

Actions:
1. ChainHashCompute(Chain_A) → hash_a.
2. ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a equals hash_b.

#### Trailing blank lines removed from subsection

Setup:
- Create file_A: `_node.md` with `## Interface` with
  trailing blank lines.
- Create file_B: same content, no trailing blank lines.

Actions:
1. ChainHashCompute(Chain_A) → hash_a.
2. ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a equals hash_b.

#### Interior blank lines preserved

Setup:
- Create file_A: `_node.md` with `## Interface` with
  blank lines between content lines.
- Create file_B: same content, interior blank lines
  removed.

Actions:
1. ChainHashCompute(Chain_A) → hash_a.
2. ChainHashCompute(Chain_B) → hash_b.

Expected: hash_a differs from hash_b.

### Target

#### Target Public and Agent both contribute

Setup:
- Create SPEC/a with `# Public` → `## Interface` and
  `# Agent` with content.
- Build Chain with target = SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Remove `# Agent` from file.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### Target without Agent — Agent skipped

Setup:
- Create SPEC/a with `# Public` → `## Interface`, no
  `# Agent`.
- Build Chain with target = SPEC/a.

Actions:
1. Call ChainHashCompute.

Expected: No error. 27-character result.

### Input

#### Input hashes full file content

Setup:
- Create an artifact file with content.
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with target = SPEC/a, input =
  ChainItem(ARTIFACT/input, file_path=<path>).

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the artifact file.
3. Call ChainHashCompute → hash_after.

Expected: hash_before differs from hash_after.

#### No input — skipped

Setup:
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with target = SPEC/a, input absent.

Actions:
1. Call ChainHashCompute.

Expected: No error. 27-character result.

### Error cases

#### Unreadable spec node file

Setup:
- Build Chain with target pointing to a non-existent
  file.

Actions:
1. Call ChainHashCompute.

Expected: Error ParseFailure.

#### Unreadable artifact file

Setup:
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with ARTIFACT dependency pointing to a
  non-existent file.

Actions:
1. Call ChainHashCompute.

Expected: Error FileUnreadable.

#### Unreadable external file

Setup:
- Create SPEC/a with `# Public` → `## Interface`.
- Build Chain with EXTERNAL dependency pointing to a
  non-existent file.

Actions:
1. Call ChainHashCompute.

Expected: Error FileUnreadable.

## Go-specific guidance

- The package name is `chainhash_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
