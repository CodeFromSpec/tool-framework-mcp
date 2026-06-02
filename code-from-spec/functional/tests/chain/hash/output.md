<!-- code-from-spec: ROOT/functional/tests/chain/hash@MnOivbwb2dExRAceSn13XKRyIdw -->

## ChainHashCompute — Test Specification

### Properties

#### Hash is deterministic

Setup: Create files on disk for a valid spec tree. Build a Chain directly.

Actions:
1. Call ChainHashCompute with the Chain.
2. Call ChainHashCompute again with the same Chain.

Expected: Both return values are identical strings.

---

#### Hash is 27 characters

Setup: Create files on disk for a valid spec tree. Build a Chain directly.

Actions:
1. Call ChainHashCompute with the Chain.

Expected: The returned string is exactly 27 characters long.

---

#### Hash changes when ancestor content changes

Setup:
- Create a ROOT node file on disk with `# Public` content.
- Create a ROOT/a node file as the target.
- Build a Chain with ancestors = [ChainItem for ROOT], target = ChainItem for ROOT/a.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify ROOT's `# Public` content on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### Hash changes when dependency content changes

Setup:
- Create ROOT, ROOT/a, ROOT/b node files on disk.
- Build a Chain with target = ChainItem for ROOT/a, dependencies = [ChainItem for ROOT/b (no qualifier)].

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify ROOT/b's `# Public` content on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### Hash changes when target Public changes

Setup:
- Create ROOT and ROOT/a node files. ROOT/a has `# Public` content.
- Build a Chain with target = ChainItem for ROOT/a.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify ROOT/a's `# Public` content on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### Hash changes when target Agent changes

Setup:
- Create ROOT and ROOT/a node files. ROOT/a has `# Agent` content.
- Build a Chain with target = ChainItem for ROOT/a.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify ROOT/a's `# Agent` content on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

### Ancestors

#### Ancestor with Public section contributes hash

Setup:
- Create ROOT node file with `# Public` content.
- Create ROOT/a node file as target.
- Build a Chain with ancestors = [ChainItem for ROOT], target = ChainItem for ROOT/a.

Actions:
1. Call ChainHashCompute.

Expected: Returns a 27-character non-empty string with no error.

---

#### Ancestor without Public section — skipped

Setup:
- Create ROOT node file with no `# Public` section (only a name section or empty body).
- Create ROOT/a node file as target (with `# Public`).
- Build Chain A with ancestors = [ChainItem for ROOT].
- Create ROOT/z node file with `# Public` content.
- Build Chain B with ancestors = [ChainItem for ROOT/z].

Actions:
1. Call ChainHashCompute with Chain A. Record result as hashA.
2. Call ChainHashCompute with Chain B. Record result as hashB.

Expected: hashA and hashB differ (the ancestor without `# Public` contributes no content).

---

#### Multiple ancestors — order matters

Setup:
- Create ROOT, ROOT/a, ROOT/a/b node files. ROOT and ROOT/a both have `# Public` content.
- Build Chain X with ancestors = [ChainItem for ROOT, ChainItem for ROOT/a] (root-first order), target = ChainItem for ROOT/a/b.
- Build Chain Y with ancestors = [ChainItem for ROOT/a, ChainItem for ROOT] (swapped order), target = ChainItem for ROOT/a/b.

Actions:
1. Call ChainHashCompute with Chain X. Record result as hashX.
2. Call ChainHashCompute with Chain Y. Record result as hashY.

Expected: hashX and hashY differ.

---

### Dependencies

#### ROOT dependency without qualifier — hashes Public

Setup:
- Create ROOT/b node file with `# Public` content.
- Build Chain with target = ChainItem for ROOT/a, dependencies = [ChainItem for ROOT/b, qualifier absent].

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify ROOT/b's `# Public` content on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### ROOT dependency with qualifier — hashes subsection

Setup:
- Create ROOT/b node file with `# Public` containing a `## Interface` subsection.
- Build Chain with dependencies = [ChainItem for ROOT/b, qualifier = "interface"].

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify the `## Interface` subsection content in ROOT/b on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### Qualifier case normalization

Setup:
- Create ROOT/b node file with `# Public` containing a `## Interface` subsection.
- Build Chain with dependencies = [ChainItem for ROOT/b, qualifier = "INTERFACE" (uppercase)].

Actions:
1. Call ChainHashCompute.

Expected: Returns a result with no error. The qualifier is normalized before matching.

---

#### ARTIFACT dependency — hashes file minus frontmatter

Setup:
- Create an artifact file on disk with frontmatter and body content.
- Build Chain with dependencies = [ChainItem for the artifact file, marked as ARTIFACT].

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify the body content of the artifact file on disk (do not change frontmatter).
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### ARTIFACT dependency — frontmatter change ignored

Setup:
- Create an artifact file on disk with frontmatter and body content.
- Build Chain with ARTIFACT dependency pointing to that file.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify only the frontmatter of the artifact file on disk (do not change body).
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 are identical.

---

#### ARTIFACT dependency — tag hash change ignored

Setup:
- Create an artifact file on disk. The body contains a line:
  `// code-from-spec: ROOT/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn`
- Build Chain with ARTIFACT dependency pointing to that file.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Change only the 27-character hash in the tag line to a different value (e.g. `zZyYxXwWvVuUtTsSrRqQpPoOnNm`).
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 are identical — the tag hash portion is neutralized before hashing.

---

### External files

#### External file — hashes all content

Setup:
- Create an external file on disk with some content.
- Build Chain with external = [FrontmatterExternal pointing to that file].

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify the external file's content on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

### Target

#### Target Public and Agent both contribute

Setup:
- Create ROOT/a node file as target with both `# Public` and `# Agent` sections.
- Build Chain with target = ChainItem for ROOT/a.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Remove the `# Agent` section from ROOT/a on disk.
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### Target without Agent — Agent skipped

Setup:
- Create ROOT/a node file as target with `# Public` only (no `# Agent` section).
- Build Chain with target = ChainItem for ROOT/a.

Actions:
1. Call ChainHashCompute.

Expected: Returns a result with no error.

---

### Input

#### Input hashes file minus frontmatter

Setup:
- Create an artifact file on disk with frontmatter and body content.
- Build Chain with input = ChainItem pointing to the artifact file.

Actions:
1. Call ChainHashCompute. Record result as hash1.
2. Modify the body content of the artifact file on disk (do not change frontmatter).
3. Call ChainHashCompute again. Record result as hash2.

Expected: hash1 and hash2 differ.

---

#### No input — skipped

Setup:
- Create a minimal spec tree on disk.
- Build Chain with input absent.

Actions:
1. Call ChainHashCompute.

Expected: Returns a result with no error.

---

### Error cases

#### Unreadable spec node file

Setup:
- Build Chain referencing a ChainItem whose file_path points to a non-existent file on disk.

Actions:
1. Call ChainHashCompute.

Expected: Error ParseFailure is returned.

---

#### Unreadable artifact file

Setup:
- Build Chain with an ARTIFACT dependency whose file_path points to a non-existent file on disk.

Actions:
1. Call ChainHashCompute.

Expected: Error FileUnreadable is returned.

---

#### Unreadable external file

Setup:
- Build Chain with external = [FrontmatterExternal pointing to a non-existent file on disk].

Actions:
1. Call ChainHashCompute.

Expected: Error FileUnreadable is returned.
