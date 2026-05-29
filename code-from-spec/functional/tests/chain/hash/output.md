<!-- code-from-spec: ROOT/functional/tests/chain/hash@gCw7U6sJxanUNJjAHHOqteINj_o -->

# Test Specification: ChainHashCompute

Tests build `Chain` records directly — they do not call `ChainResolve`.
Each test creates files on disk as needed.

---

## Properties

### Hash is deterministic

Setup:
- Create a spec node file on disk with `# Public` content.
- Build a `Chain` with that file as the target `ChainItem`.

Actions:
1. Call `ChainHashCompute` with the Chain. Record result as <hash1>.
2. Call `ChainHashCompute` again with the same Chain. Record result as <hash2>.

Expected outcome:
- <hash1> equals <hash2>.

---

### Hash is 27 characters

Setup:
- Create a spec node file on disk.
- Build a valid `Chain` with that file as the target.

Actions:
1. Call `ChainHashCompute`. Record result as <hash>.

Expected outcome:
- <hash> is exactly 27 characters long.

---

### Hash changes when ancestor content changes

Setup:
- Create a ROOT spec node file with `# Public` content "original ancestor".
- Create a ROOT/a spec node file.
- Build a `Chain` with ancestors = [ChainItem for ROOT], target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify ROOT's `# Public` content on disk to "modified ancestor".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### Hash changes when dependency content changes

Setup:
- Create ROOT, ROOT/a, ROOT/b spec node files on disk.
- ROOT/b has `# Public` content "original dependency".
- Build a `Chain` with target = ChainItem for ROOT/a, dependencies = [ChainItem for ROOT/b (no qualifier)].

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify ROOT/b's `# Public` content on disk to "modified dependency".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### Hash changes when target Public changes

Setup:
- Create ROOT and ROOT/a spec node files on disk.
- ROOT/a has `# Public` content "original public".
- Build a `Chain` with target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify ROOT/a's `# Public` content on disk to "modified public".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### Hash changes when target Agent changes

Setup:
- Create ROOT and ROOT/a spec node files on disk.
- ROOT/a has `# Agent` content "original agent".
- Build a `Chain` with target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify ROOT/a's `# Agent` content on disk to "modified agent".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

## Ancestors

### Ancestor with Public section contributes hash

Setup:
- Create a ROOT spec node file with `# Public` content "ancestor public content".
- Create a ROOT/a spec node file.
- Build a `Chain` with ancestors = [ChainItem for ROOT], target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute`. Record result as <hash>.

Expected outcome:
- No error is raised.
- <hash> is exactly 27 characters long and is non-empty.

---

### Ancestor without Public section — skipped

Setup:
- Create a ROOT spec node file with only a name section — no `# Public` section.
- Create a ROOT/a spec node file.
- Build Chain A with ancestors = [ChainItem for ROOT], target = ChainItem for ROOT/a.
- Create a second ROOT spec node file with a `# Public` section.
- Build Chain B with ancestors = [ChainItem for that ROOT], target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute` with Chain A. Record result as <hashA>.
2. Call `ChainHashCompute` with Chain B. Record result as <hashB>.

Expected outcome:
- <hashA> does not equal <hashB>.

---

### Multiple ancestors — order matters

Setup:
- Create ROOT, ROOT/a, ROOT/a/b spec node files on disk.
- ROOT has `# Public` content "root public".
- ROOT/a has `# Public` content "a public".
- Build Chain A with ancestors = [ChainItem for ROOT, ChainItem for ROOT/a] (root-first order), target = ChainItem for ROOT/a/b.
- Build Chain B with ancestors = [ChainItem for ROOT/a, ChainItem for ROOT] (reversed order), target = ChainItem for ROOT/a/b.

Actions:
1. Call `ChainHashCompute` with Chain A. Record result as <hashA>.
2. Call `ChainHashCompute` with Chain B. Record result as <hashB>.

Expected outcome:
- <hashA> does not equal <hashB>.

---

## Dependencies

### ROOT dependency without qualifier — hashes Public

Setup:
- Create ROOT/b spec node file with `# Public` content "b public original".
- Build a `Chain` with target = ChainItem for ROOT/b, dependencies = [ChainItem for ROOT/b, qualifier absent].

  Note: target and dependency may be separate nodes in practice; use a minimal valid Chain.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify ROOT/b's `# Public` content on disk to "b public modified".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### ROOT dependency with qualifier — hashes subsection

Setup:
- Create ROOT/b spec node file with `# Public` containing `## Interface` subsection with content "interface original".
- Build a `Chain` with a dependency ChainItem for ROOT/b, qualifier = "interface".

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify the `## Interface` subsection content on disk to "interface modified".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### Qualifier case normalization

Setup:
- Create ROOT/b spec node file with `# Public` containing `## Interface` subsection.
- Build a `Chain` with a dependency ChainItem for ROOT/b, qualifier = "INTERFACE" (uppercase).

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- No error is raised.
- Result is exactly 27 characters long.

---

### ARTIFACT dependency — hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter block and body content "original body".
- Build a `Chain` with a dependency ChainItem pointing to that artifact file (ARTIFACT reference).

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify the body content of the artifact file to "modified body".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### ARTIFACT dependency — frontmatter change ignored

Setup:
- Create an artifact file with frontmatter and body content "stable body".
- Build a `Chain` with a dependency ChainItem pointing to that artifact file (ARTIFACT reference).

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify only the frontmatter of the artifact file (body unchanged).
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> equals <hash2>.

---

## External Files

### External whole file — hashes all content

Setup:
- Create an external file with content "external original".
- Build a `Chain` with an external entry pointing to that file, no fragments.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify the external file content to "external modified".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### External with fragments — hashes declared ranges

Setup:
- Create an external file with 10 lines (line 1 through line 10).
- Build a `Chain` with an external entry for that file, fragments = [{lines: "3-5"}].

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify line 4 of the external file.
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### External with fragments — change outside range ignored

Setup:
- Create an external file with 10 lines.
- Build a `Chain` with an external entry for that file, fragments = [{lines: "3-5"}].

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify line 8 of the external file (outside range 3-5).
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> equals <hash2>.

---

### External with multiple fragments — declaration order

Setup:
- Create an external file with 10 lines.
- Build Chain A with an external entry for that file, fragments = [{lines: "6-8"}, {lines: "1-3"}].
- Build Chain B with an external entry for that file, fragments = [{lines: "1-3"}, {lines: "6-8"}] (reversed order).

Actions:
1. Call `ChainHashCompute` with Chain A. Record result as <hashA>.
2. Call `ChainHashCompute` with Chain B. Record result as <hashB>.

Expected outcome:
- <hashA> does not equal <hashB>.

---

## Target

### Target Public and Agent both contribute

Setup:
- Create ROOT/a spec node file with both `# Public` content "target public" and `# Agent` content "target agent".
- Build a `Chain` with target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Remove the `# Agent` section from ROOT/a's file on disk.
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### Target without Agent — Agent skipped

Setup:
- Create ROOT/a spec node file with `# Public` content only — no `# Agent` section.
- Build a `Chain` with target = ChainItem for ROOT/a.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- No error is raised.
- Result is exactly 27 characters long.

---

## Input

### Input hashes file minus frontmatter

Setup:
- Create an artifact file with frontmatter and body content "input body original".
- Build a `Chain` with input = ChainItem pointing to that artifact file.

Actions:
1. Call `ChainHashCompute`. Record result as <hash1>.
2. Modify the body content of the artifact file to "input body modified".
3. Call `ChainHashCompute` again. Record result as <hash2>.

Expected outcome:
- <hash1> does not equal <hash2>.

---

### No input — skipped

Setup:
- Create a spec node file on disk.
- Build a `Chain` with input absent.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- No error is raised.
- Result is exactly 27 characters long.

---

## Error Cases

### Unreadable spec node file

Setup:
- Build a `Chain` whose target `ChainItem` references a spec node file path that does not exist on disk.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- Raises error "parse failure".

---

### Unreadable artifact file

Setup:
- Build a `Chain` with an ARTIFACT dependency `ChainItem` whose file path does not exist on disk.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- Raises error "file unreadable".

---

### Unreadable external file

Setup:
- Build a `Chain` with an external entry whose file path does not exist on disk.

Actions:
1. Call `ChainHashCompute`.

Expected outcome:
- Raises error "file unreadable".
