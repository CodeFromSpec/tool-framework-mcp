<!-- code-from-spec: ROOT/functional/tests/chain/hash@c9JNpHP_fGApBQbC5WyCiwxh-w0 -->

## Test suite: ChainHashCompute

---

### Properties

---

#### Hash is deterministic

Setup:
- Create a spec tree with ROOT containing `# Public` content and ROOT/a as target.
- Build a Chain record with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute with the Chain.
- Call ChainHashCompute again with the same Chain.

Expected outcome:
- Both results are identical strings.

---

#### Hash is 27 characters

Setup:
- Create a spec tree with a valid Chain (any structure).

Actions:
- Call ChainHashCompute with the Chain.

Expected outcome:
- The result is exactly 27 characters long.

---

#### Hash changes when ancestor content changes

Setup:
- Create ROOT with `# Public` section containing some content.
- Create ROOT/a as target node.
- Build a Chain with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify ROOT's `# Public` content on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### Hash changes when dependency content changes

Setup:
- Create ROOT as ancestor node.
- Create ROOT/b with `# Public` content.
- Create ROOT/a as target, with a dependency on ROOT/b (no qualifier).
- Build a Chain with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem, dependencies = [ROOT/b ChainItem].

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify ROOT/b's `# Public` content on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### Hash changes when target Public changes

Setup:
- Create ROOT as ancestor node.
- Create ROOT/a as target with `# Public` content.
- Build a Chain with target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify ROOT/a's `# Public` content on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### Hash changes when target Agent changes

Setup:
- Create ROOT as ancestor node.
- Create ROOT/a as target with `# Public` content and `# Agent` content.
- Build a Chain with target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify ROOT/a's `# Agent` content on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

### Ancestors

---

#### Ancestor with Public section contributes hash

Setup:
- Create ROOT with `# Public` content.
- Create ROOT/a as target.
- Build a Chain with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute.

Expected outcome:
- Result is a non-empty string of exactly 27 characters.

---

#### Ancestor without Public section — skipped

Setup:
- Create ROOT with no `# Public` section (only a name section or empty content).
- Create ROOT/a as target with `# Public` content.
- Build Chain A with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem.
- Create ROOT/c with `# Public` content.
- Create ROOT/a2 as target (same content as ROOT/a).
- Build Chain B with ancestors = [ROOT/c ChainItem], target = ROOT/a2 ChainItem.

Actions:
- Call ChainHashCompute with Chain A. Record result as hashA.
- Call ChainHashCompute with Chain B. Record result as hashB.

Expected outcome:
- hashA and hashB differ, because ROOT without `# Public` contributes nothing to the hash while ROOT/c with `# Public` does.

---

#### Multiple ancestors — order matters

Setup:
- Create ROOT with `# Public` content "root content".
- Create ROOT/a with `# Public` content "a content".
- Create ROOT/a/b as target.
- Build Chain Forward with ancestors = [ROOT ChainItem, ROOT/a ChainItem] (root-first), target = ROOT/a/b ChainItem.
- Build Chain Reversed with ancestors = [ROOT/a ChainItem, ROOT ChainItem] (reversed), target = ROOT/a/b ChainItem.

Actions:
- Call ChainHashCompute with Chain Forward. Record result as hashForward.
- Call ChainHashCompute with Chain Reversed. Record result as hashReversed.

Expected outcome:
- hashForward and hashReversed differ.

---

### Dependencies

---

#### ROOT dependency without qualifier — hashes Public

Setup:
- Create ROOT/b with `# Public` content.
- Create ROOT/a as target.
- Build a Chain with target = ROOT/a ChainItem, dependencies = [ROOT/b ChainItem with no qualifier].

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify ROOT/b's `# Public` content on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### ROOT dependency with qualifier — hashes subsection

Setup:
- Create ROOT/b with a `# Public` section containing an `## Interface` subsection.
- Create ROOT/a as target.
- Build a Chain with target = ROOT/a ChainItem, dependencies = [ROOT/b ChainItem with qualifier = "interface"].

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify the `## Interface` content inside ROOT/b's `# Public` section on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### Qualifier case normalization

Setup:
- Create ROOT/b with a `# Public` section containing an `## Interface` subsection.
- Create ROOT/a as target.
- Build a Chain with target = ROOT/a ChainItem, dependencies = [ROOT/b ChainItem with qualifier = "INTERFACE" (uppercase)].

Actions:
- Call ChainHashCompute.

Expected outcome:
- No error is raised. The qualifier is normalized and matched successfully.

---

#### ARTIFACT dependency — hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter (e.g., `---\noutput: some/path\n---`) and a body.
- Build a Chain with dependencies = [ChainItem with file_path pointing to the artifact file, no qualifier].

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify the body content of the artifact file on disk (leave frontmatter unchanged).
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### ARTIFACT dependency — frontmatter change ignored

Setup:
- Create an artifact file with frontmatter and a body.
- Build a Chain with dependencies = [ChainItem pointing to the artifact file, no qualifier].

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify only the frontmatter of the artifact file on disk (leave body unchanged).
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 are identical — frontmatter is stripped before hashing.

---

#### ARTIFACT dependency — tag hash change ignored

Setup:
- Create an artifact file with a body containing an artifact tag line:
  `// code-from-spec: ROOT/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn`
- Build a Chain with dependencies = [ChainItem pointing to that file, no qualifier].

Actions:
- Call ChainHashCompute. Record result as hash1.
- Change only the 27-character hash in the artifact tag to a different value
  (e.g., `zZyYxXwWvVuUtTsSrRqQpPoOnNm`). Leave all other content unchanged.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 are identical — the tag hash portion is neutralized before hashing.

---

### External files

---

#### External file — hashes all content

Setup:
- Create an external file with some content.
- Build a Chain with external = [FrontmatterExternal record pointing to that file], target = some target ChainItem.

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify the external file's content on disk.
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

### Target

---

#### Target Public and Agent both contribute

Setup:
- Create ROOT/a as target with both `# Public` and `# Agent` sections.
- Build a Chain with target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute. Record result as hash1.
- Remove the `# Agent` section from ROOT/a on disk (keep `# Public` unchanged).
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### Target without Agent — Agent skipped

Setup:
- Create ROOT/a as target with `# Public` only and no `# Agent` section.
- Build a Chain with target = ROOT/a ChainItem.

Actions:
- Call ChainHashCompute.

Expected outcome:
- No error is raised. The result is a 27-character string.

---

### Input

---

#### Input hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter and a body.
- Build a Chain with input = ChainItem pointing to the artifact file.

Actions:
- Call ChainHashCompute. Record result as hash1.
- Modify the body content of the artifact file on disk (leave frontmatter unchanged).
- Call ChainHashCompute again. Record result as hash2.

Expected outcome:
- hash1 and hash2 differ.

---

#### No input — skipped

Setup:
- Create ROOT/a as target with `# Public` content.
- Build a Chain with target = ROOT/a ChainItem, input = absent.

Actions:
- Call ChainHashCompute.

Expected outcome:
- No error is raised. The result is a 27-character string.

---

### Error cases

---

#### Unreadable spec node file

Setup:
- Build a Chain where the target ChainItem's file_path points to a file that does not exist on disk.

Actions:
- Call ChainHashCompute.

Expected outcome:
- Error ParseFailure is returned.

---

#### Unreadable artifact file

Setup:
- Build a Chain with a dependency ChainItem whose file_path points to a non-existent artifact file.

Actions:
- Call ChainHashCompute.

Expected outcome:
- Error FileUnreadable is returned.

---

#### Unreadable external file

Setup:
- Build a Chain with an external entry whose path points to a non-existent file.

Actions:
- Call ChainHashCompute.

Expected outcome:
- Error FileUnreadable is returned.
