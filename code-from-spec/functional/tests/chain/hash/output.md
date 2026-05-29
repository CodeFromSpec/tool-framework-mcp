<!-- code-from-spec: ROOT/functional/tests/chain/hash@gCw7U6sJxanUNJjAHHOqteINj_o -->

# ChainHashCompute — Test Cases

Each test builds a `Chain` record directly (no call to `ChainResolve`) and
creates any required files on disk before invoking `ChainHashCompute`.

---

## Properties

### Hash is deterministic

Setup:
- Create a file on disk with arbitrary content.
- Build a `Chain` with that file as the target (no ancestors, no
  dependencies, no input).

Actions:
1. Call `ChainHashCompute` with the Chain. Record result as `hash1`.
2. Call `ChainHashCompute` again with the same Chain. Record result as `hash2`.

Expected outcome:
- `hash1` equals `hash2`.

---

### Hash is 27 characters

Setup:
- Create a minimal spec file on disk.
- Build a `Chain` with that file as the target.

Actions:
1. Call `ChainHashCompute` with the Chain. Record result as `hash`.

Expected outcome:
- `hash` is exactly 27 characters long.

---

### Hash changes when ancestor content changes

Setup:
- Create a ROOT spec file with a `# Public` section containing `"original content"`.
- Create a ROOT/a spec file (the target).
- Build a `Chain` with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem.

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT's `# Public` section on disk (change to `"modified content"`).
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### Hash changes when dependency content changes

Setup:
- Create a ROOT spec file.
- Create a ROOT/a spec file (the target).
- Create a ROOT/b spec file with a `# Public` section containing `"original content"`.
- Build a `Chain` with target = ROOT/a ChainItem,
  dependencies = [ROOT/b ChainItem (no qualifier)].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/b's `# Public` section on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### Hash changes when target Public changes

Setup:
- Create a ROOT spec file.
- Create a ROOT/a spec file with a `# Public` section containing `"original content"`.
- Build a `Chain` with target = ROOT/a ChainItem.

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/a's `# Public` section on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### Hash changes when target Agent changes

Setup:
- Create a ROOT spec file.
- Create a ROOT/a spec file with a `# Agent` section containing `"original agent instructions"`.
- Build a `Chain` with target = ROOT/a ChainItem.

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/a's `# Agent` section on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

## Ancestors

### Ancestor with Public section contributes hash

Setup:
- Create a ROOT spec file with a `# Public` section containing some content.
- Create a ROOT/a spec file (the target).
- Build a `Chain` with ancestors = [ROOT ChainItem], target = ROOT/a ChainItem.

Actions:
1. Call `ChainHashCompute`. Record result as `hash`.

Expected outcome:
- `hash` is a non-empty string exactly 27 characters long.

---

### Ancestor without Public section — skipped

Setup:
- Create ROOT spec file A with no `# Public` section (only a name/title section).
- Create ROOT spec file B with a `# Public` section containing some content.
- Create a target spec file for each scenario.
- Build `Chain1` with ancestors = [ChainItem for ROOT A], target = target.
- Build `Chain2` with ancestors = [ChainItem for ROOT B], target = target (same target file).

Actions:
1. Call `ChainHashCompute` with `Chain1`. Record result as `hash1`.
2. Call `ChainHashCompute` with `Chain2`. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### Multiple ancestors — order matters

Setup:
- Create a ROOT spec file with a `# Public` section containing `"root public"`.
- Create a ROOT/a spec file with a `# Public` section containing `"a public"`.
- Create a ROOT/a/b spec file (the target).
- Build `Chain1` with ancestors = [ROOT ChainItem, ROOT/a ChainItem] (root-first),
  target = ROOT/a/b ChainItem.
- Build `Chain2` with ancestors = [ROOT/a ChainItem, ROOT ChainItem] (reversed),
  target = ROOT/a/b ChainItem.

Actions:
1. Call `ChainHashCompute` with `Chain1`. Record result as `hash1`.
2. Call `ChainHashCompute` with `Chain2`. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

## Dependencies

### ROOT dependency without qualifier — hashes Public

Setup:
- Create a ROOT/b spec file with a `# Public` section containing `"original public"`.
- Create a ROOT/a spec file (the target).
- Build a `Chain` with target = ROOT/a ChainItem,
  dependencies = [ROOT/b ChainItem, qualifier absent].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/b's `# Public` section on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### ROOT dependency with qualifier — hashes subsection

Setup:
- Create a ROOT/b spec file with a `# Public` section that contains
  a `## Interface` subsection with `"original interface content"`.
- Create a ROOT/a spec file (the target).
- Build a `Chain` with target = ROOT/a ChainItem,
  dependencies = [ROOT/b ChainItem, qualifier = `"interface"`].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the `## Interface` subsection content on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### Qualifier case normalization

Setup:
- Create a ROOT/b spec file with a `# Public` section that contains
  a `## Interface` subsection.
- Create a ROOT/a spec file (the target).
- Build a `Chain` with target = ROOT/a ChainItem,
  dependencies = [ROOT/b ChainItem, qualifier = `"INTERFACE"` (uppercase)].

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- No error is raised. The function returns a 27-character result.

---

### ARTIFACT dependency — hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter (between `---` delimiters)
  and body content `"original body"`.
- Create a target spec file.
- Build a `Chain` with target = target ChainItem,
  dependencies = [ARTIFACT ChainItem pointing to the artifact file].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify only the body content of the artifact file on disk (leave frontmatter unchanged).
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### ARTIFACT dependency — frontmatter change ignored

Setup:
- Create an artifact file with frontmatter and body content `"stable body"`.
- Create a target spec file.
- Build a `Chain` with target = target ChainItem,
  dependencies = [ARTIFACT ChainItem pointing to the artifact file].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify only the frontmatter of the artifact file on disk (leave body unchanged).
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` equals `hash2`.

---

## External files

### External whole file — hashes all content

Setup:
- Create an external file with some content.
- Create a target spec file.
- Build a `Chain` with target = target ChainItem,
  external = [FrontmatterExternal pointing to the file, no fragments].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the external file content on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### External with fragments — hashes declared ranges

Setup:
- Create an external file with exactly 10 lines of distinct content.
- Create a target spec file.
- Build a `Chain` with external = [FrontmatterExternal pointing to the file,
  fragments = [{lines: `"3-5"`}]].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify line 4 of the external file on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### External with fragments — change outside range ignored

Setup:
- Create an external file with exactly 10 lines of distinct content.
- Create a target spec file.
- Build a `Chain` with external = [FrontmatterExternal pointing to the file,
  fragments = [{lines: `"3-5"`}]].

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify line 8 of the external file on disk (outside the declared range).
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` equals `hash2`.

---

### External with multiple fragments — declaration order

Setup:
- Create an external file with exactly 10 lines of distinct content.
- Create a target spec file.
- Build `Chain1` with external = [FrontmatterExternal, fragments = [{lines: `"6-8"`}, {lines: `"1-3"`}]].
- Build `Chain2` with external = [FrontmatterExternal, fragments = [{lines: `"1-3"`}, {lines: `"6-8"`}]]
  (fragments in reversed order, same file).

Actions:
1. Call `ChainHashCompute` with `Chain1`. Record result as `hash1`.
2. Call `ChainHashCompute` with `Chain2`. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

## Target

### Target Public and Agent both contribute

Setup:
- Create a ROOT spec file.
- Create a ROOT/a spec file with both a `# Public` section and a `# Agent` section.
- Build a `Chain` with target = ROOT/a ChainItem.

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Remove the `# Agent` section from the ROOT/a file on disk.
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### Target without Agent — Agent skipped

Setup:
- Create a ROOT spec file.
- Create a ROOT/a spec file with a `# Public` section only (no `# Agent` section).
- Build a `Chain` with target = ROOT/a ChainItem.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- No error is raised. The function returns a 27-character result.

---

## Input

### Input hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter and body content `"original body"`.
- Create a target spec file.
- Build a `Chain` with target = target ChainItem,
  input = ChainItem pointing to the artifact file.

Actions:
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the body content of the artifact file on disk (leave frontmatter unchanged).
3. Call `ChainHashCompute` again. Record result as `hash2`.

Expected outcome:
- `hash1` does not equal `hash2`.

---

### No input — skipped

Setup:
- Create a target spec file.
- Build a `Chain` with target = target ChainItem, input = absent.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- No error is raised. The function returns a 27-character result.

---

## Error cases

### Unreadable spec node file

Setup:
- Build a `Chain` where the target `ChainItem` references a file path
  that does not exist on disk.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- Raises error `"parse failure"`.

---

### Unreadable artifact file

Setup:
- Create a target spec file.
- Build a `Chain` with target = target ChainItem,
  dependencies = [ARTIFACT ChainItem pointing to a file path that does not exist on disk].

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- Raises error `"file unreadable"`.

---

### Unreadable external file

Setup:
- Create a target spec file.
- Build a `Chain` with target = target ChainItem,
  external = [FrontmatterExternal pointing to a file path that does not exist on disk].

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- Raises error `"file unreadable"`.
