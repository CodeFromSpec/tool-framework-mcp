<!-- code-from-spec: ROOT/functional/tests/chain/hash@lmmdCKJyIRpqEPGzuQkaE5Wyhac -->

# ChainHashCompute — Test Specification

## Properties

---

### Hash is deterministic

Setup:
- Create a minimal spec tree on disk (e.g., a ROOT node file with some content).
- Build a `Chain` record with `target` pointing to that node.

Actions:
- Call `ChainHashCompute(chain)` → result_1.
- Call `ChainHashCompute(chain)` again → result_2.

Expected outcome:
- result_1 equals result_2.

---

### Hash is 27 characters

Setup:
- Create a minimal spec tree on disk.
- Build a valid `Chain` record.

Actions:
- Call `ChainHashCompute(chain)` → result.

Expected outcome:
- `len(result)` equals 27.

---

### Hash changes when ancestor content changes

Setup:
- Create two node files on disk: ROOT (with `# Public` section containing some content) and ROOT/a (target, minimal content).
- Build a `Chain` record with `ancestors = [ChainItem(ROOT)]` and `target = ChainItem(ROOT/a)`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify ROOT's `# Public` section content on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### Hash changes when dependency content changes

Setup:
- Create three node files on disk: ROOT, ROOT/a (target), ROOT/b (dependency, with `# Public` content).
- Build a `Chain` with `target = ChainItem(ROOT/a)` and `dependencies = [ChainItem(ROOT/b)]`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify ROOT/b's `# Public` content on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### Hash changes when target Public changes

Setup:
- Create node files: ROOT and ROOT/a (target, with `# Public` content).
- Build a `Chain` with `target = ChainItem(ROOT/a)`, no ancestors, no dependencies.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify ROOT/a's `# Public` section on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### Hash changes when target Agent changes

Setup:
- Create node files: ROOT and ROOT/a (target, with `# Agent` content).
- Build a `Chain` with `target = ChainItem(ROOT/a)`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify ROOT/a's `# Agent` section on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

## Ancestors

---

### Ancestor with Public section contributes hash

Setup:
- Create ROOT node file with a `# Public` section containing some content.
- Create ROOT/a node file as target.
- Build a `Chain` with `ancestors = [ChainItem(ROOT)]`, `target = ChainItem(ROOT/a)`.

Actions:
- Call `ChainHashCompute(chain)` → result.

Expected outcome:
- No error.
- `len(result)` equals 27.

---

### Ancestor without Public section — skipped

Setup:
- Create ROOT node file with no `# Public` section (only a name/title section or empty body).
- Create ROOT/a node file as target.
- Build chain_no_public with `ancestors = [ChainItem(ROOT, no public)]`, `target = ChainItem(ROOT/a)`.
- Create a second ROOT node file that has a `# Public` section.
- Build chain_with_public with `ancestors = [ChainItem(ROOT, with public)]`, `target = ChainItem(ROOT/a)`.

Actions:
- Call `ChainHashCompute(chain_no_public)` → hash_no_public.
- Call `ChainHashCompute(chain_with_public)` → hash_with_public.

Expected outcome:
- hash_no_public does not equal hash_with_public.

---

### Multiple ancestors — order matters

Setup:
- Create node files on disk: ROOT (with `# Public`), ROOT/a (with `# Public`), ROOT/a/b (target).
- Build chain_forward with `ancestors = [ChainItem(ROOT), ChainItem(ROOT/a)]` (root-first order).
- Build chain_reversed with `ancestors = [ChainItem(ROOT/a), ChainItem(ROOT)]` (swapped order).
- Both chains have `target = ChainItem(ROOT/a/b)`.

Actions:
- Call `ChainHashCompute(chain_forward)` → hash_forward.
- Call `ChainHashCompute(chain_reversed)` → hash_reversed.

Expected outcome:
- hash_forward does not equal hash_reversed.

---

## Dependencies

---

### ROOT dependency without qualifier — hashes Public

Setup:
- Create a node file ROOT/b with `# Public` content.
- Build a `Chain` with `target = ChainItem(ROOT/a)` and `dependencies = [ChainItem(ROOT/b, qualifier=absent)]`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify ROOT/b's `# Public` content on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### ROOT dependency with qualifier — hashes subsection

Setup:
- Create ROOT/b node file with `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain` with `dependencies = [ChainItem(ROOT/b, qualifier="interface")]`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify the `## Interface` subsection content in ROOT/b on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### Qualifier case normalization

Setup:
- Create ROOT/b node file with `# Public` section containing a `## Interface` subsection.
- Build a `Chain` with `dependencies = [ChainItem(ROOT/b, qualifier="INTERFACE")]` (uppercase).

Actions:
- Call `ChainHashCompute(chain)` → result.

Expected outcome:
- No error is raised.
- `len(result)` equals 27.

---

### ARTIFACT dependency — hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter (e.g., a YAML block delimited by `---`) and body content below the frontmatter.
- Build a `Chain` with a dependency `ChainItem` pointing to that artifact file path (ARTIFACT type, qualifier=absent).

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify only the body content of the artifact file on disk (leave frontmatter unchanged).
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### ARTIFACT dependency — frontmatter change ignored

Setup:
- Create an artifact file with frontmatter and body content.
- Build a `Chain` with a dependency `ChainItem` pointing to that artifact file (ARTIFACT type).

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify only the frontmatter of the artifact file on disk (leave body unchanged).
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 equals hash_2.

---

## External files

---

### External whole file — hashes all content

Setup:
- Create an external file with some content.
- Build a `Chain` with `external = [FrontmatterExternal(path=<file>, fragments=absent)]`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify the external file's content on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### External with fragments — hashes declared ranges

Setup:
- Create an external file with exactly 10 lines of distinct content.
- Build a `Chain` with `external = [FrontmatterExternal(path=<file>, fragments=[{lines: "3-5"}])]`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify line 4 of the external file on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### External with fragments — change outside range ignored

Setup:
- Create an external file with exactly 10 lines of distinct content.
- Build a `Chain` with `external = [FrontmatterExternal(path=<file>, fragments=[{lines: "3-5"}])]`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify line 8 of the external file on disk (outside the declared range).
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 equals hash_2.

---

### External with multiple fragments — declaration order

Setup:
- Create an external file with exactly 10 lines of distinct content.
- Build chain_order_a with `external = [FrontmatterExternal(path=<file>, fragments=[{lines: "6-8"}, {lines: "1-3"}])]`.
- Build chain_order_b with `external = [FrontmatterExternal(path=<file>, fragments=[{lines: "1-3"}, {lines: "6-8"}])]`.

Actions:
- Call `ChainHashCompute(chain_order_a)` → hash_a.
- Call `ChainHashCompute(chain_order_b)` → hash_b.

Expected outcome:
- hash_a does not equal hash_b.

---

## Target

---

### Target Public and Agent both contribute

Setup:
- Create ROOT/a node file with both `# Public` and `# Agent` sections containing content.
- Build a `Chain` with `target = ChainItem(ROOT/a)`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Remove the `# Agent` section from ROOT/a on disk.
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### Target without Agent — Agent skipped

Setup:
- Create ROOT/a node file with only `# Public` content (no `# Agent` section).
- Build a `Chain` with `target = ChainItem(ROOT/a)`.

Actions:
- Call `ChainHashCompute(chain)` → result.

Expected outcome:
- No error is raised.
- `len(result)` equals 27.

---

## Input

---

### Input hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter and body content.
- Build a `Chain` with `input = ChainItem(path=<artifact file>)`.

Actions:
- Call `ChainHashCompute(chain)` → hash_1.
- Modify the body content of the artifact file on disk (leave frontmatter unchanged).
- Call `ChainHashCompute(chain)` → hash_2.

Expected outcome:
- hash_1 does not equal hash_2.

---

### No input — skipped

Setup:
- Create a minimal spec tree on disk.
- Build a `Chain` with `input = absent`.

Actions:
- Call `ChainHashCompute(chain)` → result.

Expected outcome:
- No error is raised.
- `len(result)` equals 27.

---

## Error cases

---

### Unreadable spec node file

Setup:
- Build a `Chain` with `target = ChainItem` referencing a spec node file path that does not exist on disk.

Actions:
- Call `ChainHashCompute(chain)`.

Expected outcome:
- Error `ParseFailure` is raised.

---

### Unreadable artifact file

Setup:
- Build a `Chain` with a dependency `ChainItem` of ARTIFACT type pointing to a file path that does not exist on disk.

Actions:
- Call `ChainHashCompute(chain)`.

Expected outcome:
- Error `FileUnreadable` is raised.

---

### Unreadable external file

Setup:
- Build a `Chain` with `external = [FrontmatterExternal(path=<non-existent file>)]`.

Actions:
- Call `ChainHashCompute(chain)`.

Expected outcome:
- Error `FileUnreadable` is raised.
