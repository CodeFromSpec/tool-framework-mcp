<!-- code-from-spec: ROOT/functional/tests/chain/hash@vNHYVvog2_FG2s4RGm_YpV32_ag -->

# Public

## Test cases

### Properties

#### Hash is deterministic

Setup: Create spec node files on disk for a minimal
chain (ROOT and ROOT/a as target).

Actions: Build a Chain with target = ROOT/a. Call
ChainHashCompute(chain). Call ChainHashCompute(chain)
a second time with the same Chain.

Expected: Both results are identical strings.

---

#### Hash is 27 characters

Setup: Create spec node files on disk for a minimal
chain (ROOT and ROOT/a as target).

Actions: Build a Chain with target = ROOT/a. Call
ChainHashCompute(chain).

Expected: The result is exactly 27 characters long.

---

#### Hash changes when ancestor content changes

Setup: Create ROOT with `# Public` containing line
"original content". Create ROOT/a as target with
`# Public`. Build a Chain with ancestors = [ROOT],
target = ROOT/a.

Actions: Call ChainHashCompute(chain). Store result
as hash1. Modify ROOT's `# Public` content on disk
to "modified content". Call ChainHashCompute(chain)
again. Store result as hash2.

Expected: hash1 and hash2 differ.

---

#### Hash changes when dependency content changes

Setup: Create ROOT, ROOT/a as target, ROOT/b as a
spec node dependency with `# Public` content. Build
Chain with target = ROOT/a, dependencies = [ROOT/b].

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify ROOT/b's `# Public` content on disk. Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### Hash changes when target Public changes

Setup: Create ROOT, ROOT/a as target with `# Public`
containing "original". Build Chain with target = ROOT/a.

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify ROOT/a's `# Public` content to "changed". Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### Hash changes when target Agent changes

Setup: Create ROOT, ROOT/a as target with `# Public`
and `# Agent` containing "original agent". Build
Chain with target = ROOT/a.

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify ROOT/a's `# Agent` content to "changed agent".
Call ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

### Ancestors

#### Ancestor with Public section contributes hash

Setup: Create ROOT with `# Public` containing
"ancestor content". Create ROOT/a as target with
`# Public`.

Actions: Build Chain with ancestors = [ROOT],
target = ROOT/a. Call ChainHashCompute(chain).

Expected: Result is a non-empty 27-character string.

---

#### Ancestor without Public section — skipped

Setup: Create ROOT with only a name section (no
`# Public`). Create ROOT/a as target with `# Public`.

Actions: Build Chain A with ancestors = [ROOT],
target = ROOT/a. Call ChainHashCompute(chain A).
Store as hash_no_public.

Create ROOT2 with `# Public` containing content.
Build Chain B with ancestors = [ROOT2], target = ROOT/a
(same target file). Call ChainHashCompute(chain B).
Store as hash_with_public.

Expected: hash_no_public and hash_with_public differ.

---

#### Multiple ancestors — order matters

Setup: Create ROOT with `# Public` content "root".
Create ROOT/a with `# Public` content "mid". Create
ROOT/a/b as target with `# Public`.

Actions: Build Chain X with ancestors = [ROOT, ROOT/a]
(root-first order), target = ROOT/a/b. Call
ChainHashCompute(chain X). Store as hash_natural.

Build Chain Y with ancestors = [ROOT/a, ROOT] (reversed
order), target = ROOT/a/b. Call
ChainHashCompute(chain Y). Store as hash_reversed.

Expected: hash_natural and hash_reversed differ.

---

### Dependencies

#### ROOT dependency without qualifier — hashes Public

Setup: Create ROOT, ROOT/a as target, ROOT/b with
`# Public` containing "dep content". Build Chain with
target = ROOT/a, dependencies = [ChainItem(ROOT/b,
qualifier absent)].

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify ROOT/b's `# Public` content. Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### ROOT dependency with qualifier — hashes subsection

Setup: Create ROOT, ROOT/a as target, ROOT/b with
`# Public` containing `## Interface` subsection with
content "original interface". Build Chain with
target = ROOT/a, dependencies = [ChainItem(ROOT/b,
qualifier = "interface")].

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify the `## Interface` content in ROOT/b. Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### Qualifier case normalization

Setup: Create ROOT, ROOT/a as target, ROOT/b with
`# Public` containing `## Interface` subsection. Build
Chain with target = ROOT/a, dependencies = [ChainItem(
ROOT/b, qualifier = "INTERFACE")].

Actions: Call ChainHashCompute(chain).

Expected: No error is raised. Result is a 27-character
string.

---

#### ARTIFACT dependency — hashes file minus frontmatter

Setup: Create an artifact file with frontmatter block
(between "---" delimiters) and body content "body line".
Build Chain with target = ROOT/a, dependencies =
[ChainItem(ARTIFACT/some/node, file_path pointing to
the artifact file)].

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify only the body content of the artifact file.
Call ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### ARTIFACT dependency — frontmatter change ignored

Setup: Create an artifact file with frontmatter and
body content "body line". Build Chain with ARTIFACT
dependency pointing to that file.

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify only the frontmatter of the artifact file (not
the body). Call ChainHashCompute(chain). Store as hash2.

Expected: hash1 equals hash2.

---

### External files

#### External file — hashes all content

Setup: Create an external file with content "external
content". Build Chain with external = [entry pointing
to the file], target = ROOT/a.

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify the external file's content. Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

### Target

#### Target Public and Agent both contribute

Setup: Create ROOT, ROOT/a as target with `# Public`
content "pub" and `# Agent` content "agent". Build
Chain with target = ROOT/a.

Actions: Call ChainHashCompute(chain). Store as hash1.
Remove `# Agent` from ROOT/a on disk. Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### Target without Agent — Agent skipped

Setup: Create ROOT, ROOT/a as target with `# Public`
only (no `# Agent`). Build Chain with target = ROOT/a.

Actions: Call ChainHashCompute(chain).

Expected: No error is raised. Result is a 27-character
string.

---

### Input

#### Input hashes file minus frontmatter

Setup: Create an artifact file with frontmatter and
body content "input body". Build Chain with
input = ChainItem(file_path pointing to artifact file).

Actions: Call ChainHashCompute(chain). Store as hash1.
Modify the body content of the artifact file. Call
ChainHashCompute(chain). Store as hash2.

Expected: hash1 and hash2 differ.

---

#### No input — skipped

Setup: Create ROOT and ROOT/a as target. Build Chain
with input absent.

Actions: Call ChainHashCompute(chain).

Expected: No error is raised. Result is a 27-character
string.

---

### Error cases

#### Unreadable spec node file

Setup: Build a Chain where target.file_path points to
a spec node file that does not exist on disk.

Actions: Call ChainHashCompute(chain).

Expected: Error ParseFailure is raised.

---

#### Unreadable artifact file

Setup: Build a Chain with an ARTIFACT dependency whose
file_path points to a non-existent file.

Actions: Call ChainHashCompute(chain).

Expected: Error FileUnreadable is raised.

---

#### Unreadable external file

Setup: Build a Chain with an external entry whose path
points to a non-existent file.

Actions: Call ChainHashCompute(chain).

Expected: Error FileUnreadable is raised.
