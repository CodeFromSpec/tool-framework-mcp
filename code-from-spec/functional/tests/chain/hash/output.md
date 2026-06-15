<!-- code-from-spec: SPEC/functional/tests/chain/hash@isxZV_zMwigqnGUUcN1o9e5LyWc -->

## Test suite: ChainHashCompute

### Properties

#### Hash is deterministic

Setup:
- Create a `_node.md` file for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>).

Actions:
1. Call ChainHashCompute with the Chain.
2. Call ChainHashCompute again with the same Chain.

Expected outcome: Both results are identical strings.

---

#### Hash is 27 characters

Setup:
- Create a `_node.md` file for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem pointing to that file.

Actions:
1. Call ChainHashCompute with the Chain.

Expected outcome: The result is exactly 27 characters long.

---

#### Hash changes when ancestor content changes

Setup:
- Create a `_node.md` for SPEC with `# Public` containing a `## Context` subsection with initial content.
- Create a `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with ancestors = [ChainItem for SPEC], target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC's `## Context` subsection content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### Hash changes when dependency content changes

Setup:
- Create a `_node.md` for SPEC with `# Public` containing a `## Context` subsection.
- Create a `_node.md` for SPEC/b with `# Public` containing a `## Interface` subsection with initial content.
- Create a `_node.md` for SPEC/a with no extra content.
- Build a Chain with target = ChainItem for SPEC/a, dependencies = [ChainItem for SPEC/b (qualifier absent)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/b's `## Interface` subsection content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### Hash changes when target Public changes

Setup:
- Create a `_node.md` for SPEC with `# Public` containing a `## Context` subsection.
- Create a `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection with initial content.
- Build a Chain with target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/a's `## Interface` subsection content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### Hash changes when target Agent changes

Setup:
- Create a `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection and a `# Agent` section with initial content.
- Build a Chain with target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/a's `# Agent` section content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

### Ancestors

#### Ancestor with Public subsections contributes hash

Setup:
- Create a `_node.md` for SPEC with `# Public` containing a `## Context` subsection.
- Create a `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with ancestors = [ChainItem for SPEC], target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute.

Expected outcome: A non-empty result of exactly 27 characters.

---

#### Ancestor without Public section — skipped

Setup:
- Create a `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Create a `_node.md` for SPEC with `# Public` containing a `## Context` subsection.
- Build a Chain with ancestors = [ChainItem for SPEC], target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute → hash_with_public.
2. Rewrite SPEC's `_node.md` on disk to contain only a name heading (no `# Public` section).
3. Build the same Chain structure (same ChainItem records).
4. Call ChainHashCompute → hash_without_public.

Expected outcome: hash_with_public differs from hash_without_public.

---

#### Multiple ancestors — order matters

Setup:
- Create `_node.md` for SPEC with `# Public` containing a `## Context` subsection with content "root context".
- Create `_node.md` for SPEC/a with `# Public` containing a `## Context` subsection with content "a context".
- Create `_node.md` for SPEC/a/b with `# Public` containing a `## Interface` subsection.
- Build Chain_A with ancestors = [ChainItem for SPEC, ChainItem for SPEC/a], target = ChainItem for SPEC/a/b.
- Build Chain_B with ancestors = [ChainItem for SPEC/a, ChainItem for SPEC], target = ChainItem for SPEC/a/b.

Actions:
1. Call ChainHashCompute with Chain_A → hash_a.
2. Call ChainHashCompute with Chain_B → hash_b.

Expected outcome: hash_a differs from hash_b.

---

### Dependencies

#### SPEC dependency without qualifier — hashes Public subsections

Setup:
- Create `_node.md` for SPEC/b with `# Public` containing a `## Interface` subsection with initial content.
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, dependencies = [ChainItem for SPEC/b (qualifier absent)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify SPEC/b's `## Interface` subsection content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### SPEC dependency with qualifier — hashes subsection

Setup:
- Create `_node.md` for SPEC/b with `# Public` containing a `## Interface` subsection with initial content.
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, dependencies = [ChainItem(unqualified_logical_name="SPEC/b", file_path=<path>, qualifier="interface")].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the `## Interface` content of SPEC/b on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### Qualifier case normalization

Setup:
- Create `_node.md` for SPEC/b with `# Public` containing a `## Interface` subsection.
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, dependencies = [ChainItem(unqualified_logical_name="SPEC/b", file_path=<path>, qualifier="INTERFACE")].

Actions:
1. Call ChainHashCompute.

Expected outcome: No error is raised. The qualifier "INTERFACE" is normalized to "interface" before matching the subsection heading.

---

#### ARTIFACT dependency — hashes full file content

Setup:
- Create an artifact file with initial content.
- Build a Chain with target = ChainItem for some SPEC node, dependencies = [ChainItem(unqualified_logical_name="ARTIFACT/out", file_path=<path to artifact file>)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the artifact file content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### ARTIFACT dependency — tag hash change ignored

Setup:
- Create an artifact file whose body contains the line:
  `// code-from-spec: SPEC/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn`
- Build a Chain with target = ChainItem for some SPEC node, dependencies = [ChainItem(unqualified_logical_name="ARTIFACT/out", file_path=<path to artifact file>)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Change only the 27-character hash in the artifact tag line to a different value, e.g. `zZyYxXwWvVuUtTsSrRqQpPoOnNm`. Do not change any other content.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before equals hash_after — the artifact tag hash is neutralized before hashing.

---

#### EXTERNAL dependency — hashes all content

Setup:
- Create an external file with initial content.
- Build a Chain with target = ChainItem for some SPEC node, dependencies = [ChainItem(unqualified_logical_name="EXTERNAL/somefile", file_path=<path to external file>)].

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the external file content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

### Block extraction

#### Leading blank lines removed from subsection

Setup:
- Create file_A: a `_node.md` with a `## Interface` subsection where two blank lines appear between the heading and the first content line.
- Create file_B: a `_node.md` with a `## Interface` subsection where no blank lines appear between the heading and the same content line.
- Build Chain_A with target = ChainItem(file_path=<file_A>).
- Build Chain_B with target = ChainItem(file_path=<file_B>).

Actions:
1. Call ChainHashCompute with Chain_A → hash_a.
2. Call ChainHashCompute with Chain_B → hash_b.

Expected outcome: hash_a equals hash_b — leading blank lines are stripped by block extraction.

---

#### Trailing blank lines removed from subsection

Setup:
- Create file_A: a `_node.md` with a `## Interface` subsection where trailing blank lines appear after the last content line.
- Create file_B: a `_node.md` with the same `## Interface` subsection content but no trailing blank lines.
- Build Chain_A with target = ChainItem(file_path=<file_A>).
- Build Chain_B with target = ChainItem(file_path=<file_B>).

Actions:
1. Call ChainHashCompute with Chain_A → hash_a.
2. Call ChainHashCompute with Chain_B → hash_b.

Expected outcome: hash_a equals hash_b — trailing blank lines are stripped by block extraction.

---

#### Interior blank lines preserved

Setup:
- Create file_A: a `_node.md` with a `## Interface` subsection that has blank lines between content lines.
- Create file_B: a `_node.md` with the same `## Interface` subsection content but with those interior blank lines removed.
- Build Chain_A with target = ChainItem(file_path=<file_A>).
- Build Chain_B with target = ChainItem(file_path=<file_B>).

Actions:
1. Call ChainHashCompute with Chain_A → hash_a.
2. Call ChainHashCompute with Chain_B → hash_b.

Expected outcome: hash_a differs from hash_b — interior blank lines are preserved byte for byte.

---

### Target

#### Target Public and Agent both contribute

Setup:
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection and `# Agent` section with content.
- Build a Chain with target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute → hash_before.
2. Remove the `# Agent` section from SPEC/a's file on disk (keep `# Public`).
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### Target without Agent — Agent skipped

Setup:
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection and no `# Agent` section.
- Build a Chain with target = ChainItem for SPEC/a.

Actions:
1. Call ChainHashCompute.

Expected outcome: No error is raised. A 27-character result is returned.

---

### Input

#### Input hashes full file content

Setup:
- Create an artifact file with initial content.
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, input = ChainItem(unqualified_logical_name="ARTIFACT/input", file_path=<path to artifact file>).

Actions:
1. Call ChainHashCompute → hash_before.
2. Modify the artifact file content on disk.
3. Call ChainHashCompute → hash_after.

Expected outcome: hash_before differs from hash_after.

---

#### No input — skipped

Setup:
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, input absent.

Actions:
1. Call ChainHashCompute.

Expected outcome: No error is raised. A 27-character result is returned.

---

### Error cases

#### Unreadable spec node file

Setup:
- Build a Chain with target = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to a non-existent file>).

Actions:
1. Call ChainHashCompute.

Expected outcome: Error ParseFailure is raised.

---

#### Unreadable artifact file

Setup:
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, dependencies = [ChainItem(unqualified_logical_name="ARTIFACT/out", file_path=<path to a non-existent file>)].

Actions:
1. Call ChainHashCompute.

Expected outcome: Error FileUnreadable is raised.

---

#### Unreadable external file

Setup:
- Create `_node.md` for SPEC/a with `# Public` containing a `## Interface` subsection.
- Build a Chain with target = ChainItem for SPEC/a, dependencies = [ChainItem(unqualified_logical_name="EXTERNAL/somefile", file_path=<path to a non-existent file>)].

Actions:
1. Call ChainHashCompute.

Expected outcome: Error FileUnreadable is raised.
