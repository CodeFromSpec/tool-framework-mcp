<!-- code-from-spec: ROOT/functional/tests/chain/hash@vsalCLmPzVYfEDw0DkWSPpWYiC0 -->

# Test Specification: ChainHashCompute

Each test builds a `Chain` record directly (without calling `ChainResolve`),
creates any required files on disk, calls `ChainHashCompute`, and checks the
expected outcome.

---

## Properties

### Hash is deterministic

**Setup:**
- Create a spec node file at a temporary path with `# Public` content.
- Build a `Chain` with:
  - `ancestors`: empty
  - `dependencies`: empty
  - `external`: empty
  - `target`: a `ChainItem` pointing to the spec node file
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute` with the Chain. Record result as `hash1`.
2. Call `ChainHashCompute` again with the same Chain. Record result as `hash2`.

**Expected outcome:**
- `hash1` equals `hash2`.

---

### Hash is 27 characters

**Setup:**
- Create any valid spec node file on disk.
- Build a minimal `Chain` with `target` pointing to that file and all other
  fields empty or absent.

**Actions:**
1. Call `ChainHashCompute` with the Chain. Record result as `hash`.

**Expected outcome:**
- `hash` is exactly 27 characters long.

---

### Hash changes when ancestor content changes

**Setup:**
- Create a spec node file for ROOT containing a `# Public` section with content
  `"Root public content"`.
- Create a spec node file for ROOT/a with minimal content (used as target).
- Build a `Chain` with:
  - `ancestors`: [ChainItem for ROOT]
  - `dependencies`: empty
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT's spec node file: change `# Public` content to `"Modified root public content"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### Hash changes when dependency content changes

**Setup:**
- Create a spec node file for ROOT with minimal content.
- Create a spec node file for ROOT/b with `# Public` content `"Dependency content"`.
- Create a spec node file for ROOT/a (target) with `# Public` content and a
  `depends_on` reference to ROOT/b.
- Build a `Chain` with:
  - `ancestors`: [ChainItem for ROOT]
  - `dependencies`: [ChainItem for ROOT/b, qualifier absent]
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/b's spec node file: change `# Public` content to `"Modified dependency content"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### Hash changes when target Public changes

**Setup:**
- Create a spec node file for ROOT with minimal content.
- Create a spec node file for ROOT/a with `# Public` content `"Target public"`.
- Build a `Chain` with:
  - `ancestors`: [ChainItem for ROOT]
  - `dependencies`: empty
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/a's spec node file: change `# Public` content to `"Modified target public"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### Hash changes when target Agent changes

**Setup:**
- Create a spec node file for ROOT with minimal content.
- Create a spec node file for ROOT/a with both `# Public` and `# Agent`
  sections.
- Build a `Chain` with:
  - `ancestors`: [ChainItem for ROOT]
  - `dependencies`: empty
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/a's spec node file: change `# Agent` content to different text.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

## Ancestors

### Ancestor with Public section contributes hash

**Setup:**
- Create a spec node file for ROOT with a `# Public` section containing
  `"Ancestor public content"`.
- Create a spec node file for ROOT/a with minimal content (target).
- Build a `Chain` with:
  - `ancestors`: [ChainItem for ROOT]
  - `dependencies`: empty
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash`.

**Expected outcome:**
- `hash` is a non-empty string of exactly 27 characters.
- No error is returned.

---

### Ancestor without Public section — skipped

**Setup:**
- Create a spec node file for ROOT that has only a name/title section and no
  `# Public` section.
- Create a spec node file for ROOT/a with a `# Public` section (target).
- Build two Chains:
  - Chain A: `ancestors` = [ChainItem for ROOT (no Public)], `target` = ROOT/a
  - Chain B: Create a second spec node for ROOT/c with a `# Public` section.
    `ancestors` = [ChainItem for ROOT/c (has Public)], `target` = ROOT/a

**Actions:**
1. Call `ChainHashCompute` with Chain A. Record result as `hashA`.
2. Call `ChainHashCompute` with Chain B. Record result as `hashB`.

**Expected outcome:**
- `hashA` does not equal `hashB`.
  (An ancestor without `# Public` produces a different contribution than one
  with `# Public`.)

---

### Multiple ancestors — order matters

**Setup:**
- Create spec node files:
  - ROOT with `# Public` content `"Root public"`.
  - ROOT/a with `# Public` content `"A public"`.
  - ROOT/a/b as target with minimal content.
- Build two Chains:
  - Chain A: `ancestors` = [ChainItem for ROOT, ChainItem for ROOT/a]
    (root-first order), `target` = ROOT/a/b
  - Chain B: `ancestors` = [ChainItem for ROOT/a, ChainItem for ROOT]
    (reversed order), `target` = ROOT/a/b

**Actions:**
1. Call `ChainHashCompute` with Chain A. Record result as `hashA`.
2. Call `ChainHashCompute` with Chain B. Record result as `hashB`.

**Expected outcome:**
- `hashA` does not equal `hashB`.

---

## Dependencies

### ROOT dependency without qualifier — hashes Public

**Setup:**
- Create a spec node file for ROOT/b with `# Public` content `"Dep public"`.
- Create a spec node file for ROOT/a (target) with minimal content.
- Build a `Chain` with:
  - `ancestors`: empty
  - `dependencies`: [ChainItem for ROOT/b, qualifier absent]
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify ROOT/b's `# Public` content to `"Modified dep public"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### ROOT dependency with qualifier — hashes subsection

**Setup:**
- Create a spec node file for ROOT/b with `# Public` containing an
  `## Interface` subsection with content `"Interface content"`.
- Create a spec node file for ROOT/a (target) with minimal content.
- Build a `Chain` with:
  - `ancestors`: empty
  - `dependencies`: [ChainItem for ROOT/b, qualifier = "interface"]
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the `## Interface` subsection content in ROOT/b to `"Modified interface"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### Qualifier case normalization

**Setup:**
- Create a spec node file for ROOT/b with `# Public` containing an
  `## Interface` subsection.
- Create a spec node file for ROOT/a (target) with minimal content.
- Build a `Chain` with:
  - `dependencies`: [ChainItem for ROOT/b, qualifier = "INTERFACE"]
  - all other fields empty or absent
  - `target`: ChainItem for ROOT/a

**Actions:**
1. Call `ChainHashCompute`.

**Expected outcome:**
- No error is returned.
- The result is a 27-character string (the qualifier was normalized to
  match `## Interface` case-insensitively).

---

### ARTIFACT dependency — hashes file minus frontmatter

**Setup:**
- Create an artifact file with YAML frontmatter (e.g., an `outputs` field) and
  body content `"Artifact body"` after the frontmatter delimiter.
- Build a `Chain` with:
  - `dependencies`: [ChainItem pointing to the artifact file, qualifier absent]
  - all other fields empty or absent
  - `target`: a minimal spec node file

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the artifact body content to `"Modified artifact body"` (leave
   frontmatter unchanged).
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### ARTIFACT dependency — frontmatter change ignored

**Setup:**
- Create an artifact file with frontmatter and body content `"Stable body"`.
- Build a `Chain` with:
  - `dependencies`: [ChainItem pointing to the artifact file, qualifier absent]
  - all other fields empty or absent
  - `target`: a minimal spec node file

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify only the frontmatter of the artifact file (e.g., add or change a
   field). Leave the body unchanged.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` equals `hash2`.

---

## External files

### External whole file — hashes all content

**Setup:**
- Create an external file with content `"External file content"`.
- Build a `Chain` with:
  - `external`: [FrontmatterExternal with path pointing to the file, no fragments]
  - all other fields empty or absent
  - `target`: a minimal spec node file

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the external file content to `"Modified external content"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### External with fragments — hashes declared ranges

**Setup:**
- Create an external file with exactly 10 lines, each line containing its
  line number (e.g., `"line 1"`, `"line 2"`, ..., `"line 10"`).
- Build a `Chain` with:
  - `external`: [FrontmatterExternal with path to that file,
    fragments = [{lines: "3-5"}]]
  - all other fields empty or absent
  - `target`: a minimal spec node file

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify line 4 of the external file to `"modified line 4"`.
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### External with fragments — change outside range ignored

**Setup:**
- Create an external file with exactly 10 lines.
- Build a `Chain` with:
  - `external`: [FrontmatterExternal with path to that file,
    fragments = [{lines: "3-5"}]]
  - all other fields empty or absent
  - `target`: a minimal spec node file

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify line 8 of the external file (outside the declared range).
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` equals `hash2`.

---

### External with multiple fragments — declaration order

**Setup:**
- Create an external file with exactly 10 lines.
- Build two Chains, both with a `target` pointing to a minimal spec node file:
  - Chain A: `external` = [FrontmatterExternal with fragments =
    [{lines: "6-8"}, {lines: "1-3"}]]
  - Chain B: `external` = [FrontmatterExternal with fragments =
    [{lines: "1-3"}, {lines: "6-8"}]]

**Actions:**
1. Call `ChainHashCompute` with Chain A. Record result as `hashA`.
2. Call `ChainHashCompute` with Chain B. Record result as `hashB`.

**Expected outcome:**
- `hashA` does not equal `hashB`.

---

## Target

### Target Public and Agent both contribute

**Setup:**
- Create a spec node file for ROOT/a with both `# Public` section
  (`"Public content"`) and `# Agent` section (`"Agent content"`).
- Build a `Chain` with:
  - `ancestors`: empty
  - `dependencies`: empty
  - `external`: empty
  - `target`: ChainItem for ROOT/a
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Remove the `# Agent` section from ROOT/a's file (keep `# Public` only).
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### Target without Agent — Agent skipped

**Setup:**
- Create a spec node file for ROOT/a with `# Public` content only (no
  `# Agent` section).
- Build a `Chain` with `target` pointing to ROOT/a and all other fields empty
  or absent.

**Actions:**
1. Call `ChainHashCompute`.

**Expected outcome:**
- No error is returned.
- The result is a 27-character string.

---

## Input

### Input hashes file minus frontmatter

**Setup:**
- Create an artifact file with YAML frontmatter and body content
  `"Input body content"`.
- Build a `Chain` with:
  - `ancestors`: empty
  - `dependencies`: empty
  - `external`: empty
  - `target`: a minimal spec node file
  - `input`: ChainItem pointing to the artifact file

**Actions:**
1. Call `ChainHashCompute`. Record result as `hash1`.
2. Modify the artifact body to `"Modified input body"` (leave frontmatter
   unchanged).
3. Call `ChainHashCompute` again. Record result as `hash2`.

**Expected outcome:**
- `hash1` does not equal `hash2`.

---

### No input — skipped

**Setup:**
- Create a minimal spec node file (target).
- Build a `Chain` with:
  - `ancestors`: empty
  - `dependencies`: empty
  - `external`: empty
  - `target`: ChainItem for the spec node file
  - `input`: absent

**Actions:**
1. Call `ChainHashCompute`.

**Expected outcome:**
- No error is returned.
- The result is a 27-character string.

---

## Error cases

### Unreadable spec node file

**Setup:**
- Build a `Chain` where `target` is a `ChainItem` whose `file_path` points to a
  file that does not exist on disk.

**Actions:**
1. Call `ChainHashCompute`.

**Expected outcome:**
- Returns error `ParseFailure`.

---

### Unreadable artifact file

**Setup:**
- Build a `Chain` with a dependency `ChainItem` whose `file_path` is an
  ARTIFACT path pointing to a file that does not exist on disk.
- `target` is a valid minimal spec node file.

**Actions:**
1. Call `ChainHashCompute`.

**Expected outcome:**
- Returns error `FileUnreadable`.

---

### Unreadable external file

**Setup:**
- Build a `Chain` with an external entry (`FrontmatterExternal`) whose path
  points to a file that does not exist on disk.
- `target` is a valid minimal spec node file.

**Actions:**
1. Call `ChainHashCompute`.

**Expected outcome:**
- Returns error `FileUnreadable`.
