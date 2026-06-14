<!-- code-from-spec: ROOT/functional/tests/chain/hash@jd6ZQNUM27i9Li-TirzPTZMi7cM -->

# Test Specification: ChainHashCompute

---

## Properties

### Hash is deterministic

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC` with a `# Public` section containing a `## Context` subsection with some content.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain` directly:
  - `ancestors` = [ChainItem(unqualified_logical_name="SPEC", file_path=<path to SPEC/_node.md>, qualifier=absent)]
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `result1`.
- Call `ChainHashCompute(chain)` → `result2`.

Expected outcome:
- `result1` equals `result2`.

---

### Hash is 27 characters

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `result`.

Expected outcome:
- `len(result)` equals 27.

---

### Hash changes when ancestor content changes

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC` with a `# Public` section containing:
  ```
  ## Context
  original ancestor content
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = [ChainItem(unqualified_logical_name="SPEC", file_path=<path to SPEC/_node.md>, qualifier=absent)]
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/_node.md` so the `## Context` subsection reads `modified ancestor content`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### Hash changes when dependency content changes

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC` with a `# Public` section containing a `## Context` subsection with some content.
- Write `_node.md` for `SPEC/b` with a `# Public` section containing:
  ```
  ## Interface
  original dependency content
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="SPEC/b", file_path=<path to SPEC/b/_node.md>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/b/_node.md` so the `## Interface` subsection reads `modified dependency content`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### Hash changes when target Public changes

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing:
  ```
  ## Interface
  original target interface
  ```
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/a/_node.md` so the `## Interface` subsection reads `modified target interface`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### Hash changes when target Agent changes

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/a` with:
  - A `# Public` section containing a `## Interface` subsection with some content.
  - A `# Agent` section with `original agent content`.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/a/_node.md` so `# Agent` reads `modified agent content`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

## Ancestors

### Ancestor with Public subsections contributes hash

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC` with a `# Public` section containing:
  ```
  ## Context
  some ancestor context
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = [ChainItem(unqualified_logical_name="SPEC", file_path=<path to SPEC/_node.md>, qualifier=absent)]
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `result`.

Expected outcome:
- No error.
- `len(result)` equals 27.

---

### Ancestor without Public section — skipped

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC` with a `# Public` section containing:
  ```
  ## Context
  some context
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = [ChainItem(unqualified_logical_name="SPEC", file_path=<path to SPEC/_node.md>, qualifier=absent)]
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash_with_public`.
- Overwrite `SPEC/_node.md` with content that has only a name heading and no `# Public` section (e.g., `# SPEC`).
- Build the same `Chain` structure with the same ChainItem values.
- Call `ChainHashCompute(chain)` → `hash_without_public`.

Expected outcome:
- `hash_with_public` differs from `hash_without_public`.

---

### Multiple ancestors — order matters

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC` with a `# Public` section containing:
  ```
  ## Context
  root context
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing:
  ```
  ## Context
  mid context
  ```
- Write `_node.md` for `SPEC/a/b` with a `# Public` section containing a `## Interface` subsection with some content.
- Build `chain_forward`:
  - `ancestors` = [ChainItem("SPEC", <path SPEC/_node.md>), ChainItem("SPEC/a", <path SPEC/a/_node.md>)]
  - `dependencies` = []
  - `target` = ChainItem("SPEC/a/b", <path SPEC/a/b/_node.md>)
  - `input` = absent
- Build `chain_reversed`:
  - `ancestors` = [ChainItem("SPEC/a", <path SPEC/a/_node.md>), ChainItem("SPEC", <path SPEC/_node.md>)]
  - `dependencies` = []
  - `target` = ChainItem("SPEC/a/b", <path SPEC/a/b/_node.md>)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain_forward)` → `hash_forward`.
- Call `ChainHashCompute(chain_reversed)` → `hash_reversed`.

Expected outcome:
- `hash_forward` differs from `hash_reversed`.

---

## Dependencies

### SPEC dependency without qualifier — hashes Public subsections

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/b` with a `# Public` section containing:
  ```
  ## Interface
  original interface
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="SPEC/b", file_path=<path to SPEC/b/_node.md>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/b/_node.md` so `## Interface` reads `modified interface`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### SPEC dependency with qualifier — hashes subsection

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/b` with a `# Public` section containing:
  ```
  ## Interface
  original interface
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="SPEC/b", file_path=<path to SPEC/b/_node.md>, qualifier="interface")]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/b/_node.md` so `## Interface` reads `modified interface`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### Qualifier case normalization

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/b` with a `# Public` section containing:
  ```
  ## Interface
  some interface content
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="SPEC/b", file_path=<path to SPEC/b/_node.md>, qualifier="INTERFACE")]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `result`.

Expected outcome:
- No error.
- `len(result)` equals 27.

---

### ARTIFACT dependency — hashes full file content

Setup:
- Create a temp directory.
- Create an artifact file (e.g., `output.md`) with content `original artifact content`.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="ARTIFACT/x", file_path=<path to artifact file>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite the artifact file with `modified artifact content`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### ARTIFACT dependency — tag hash change ignored

Setup:
- Create a temp directory.
- Create an artifact file with body content:
  ```
  // code-from-spec: SPEC/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn
  some body content
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="ARTIFACT/x", file_path=<path to artifact file>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite the artifact file, changing only the 27-character hash in the tag to `zZyYxXwWvVuUtTsSrRqQpPoOnNm`:
  ```
  // code-from-spec: SPEC/x/y@zZyYxXwWvVuUtTsSrRqQpPoOnNm
  some body content
  ```
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` equals `hash2`.

---

### EXTERNAL dependency — hashes all content

Setup:
- Create a temp directory.
- Create an external file with content `original external content`.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="EXTERNAL/rules.md", file_path=<path to external file>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite the external file with `modified external content`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

## Block Extraction

### Leading blank lines removed from subsection

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/b` — version A — with a `# Public` section where `## Interface` has two blank lines before the first content line:
  ```
  # Public

  ## Interface


  interface content line
  ```
- Write `_node.md` for `SPEC/b` — version B — with the same `## Interface` subsection but no blank lines before the content:
  ```
  # Public

  ## Interface
  interface content line
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build `chain_a` using version A for `SPEC/b`:
  - `ancestors` = []
  - `dependencies` = [ChainItem("SPEC/b", <path to version A _node.md>, qualifier=absent)]
  - `target` = ChainItem("SPEC/a", <path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Write version A to `SPEC/b/_node.md`.
- Call `ChainHashCompute(chain_a)` → `hash_a`.
- Overwrite `SPEC/b/_node.md` with version B.
- Build `chain_b` using the same ChainItem for `SPEC/b` (same path).
- Call `ChainHashCompute(chain_b)` → `hash_b`.

Expected outcome:
- `hash_a` equals `hash_b`.

---

### Trailing blank lines removed from subsection

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/b` — version A — with a `# Public` section where `## Interface` has trailing blank lines after the last content line:
  ```
  # Public

  ## Interface
  interface content line


  ```
- Write `_node.md` for `SPEC/b` — version B — with the same content but no trailing blank lines:
  ```
  # Public

  ## Interface
  interface content line
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.

Actions:
- Write version A to `SPEC/b/_node.md`. Build `chain` pointing to it. Call `ChainHashCompute(chain)` → `hash_a`.
- Overwrite `SPEC/b/_node.md` with version B. Call `ChainHashCompute(chain)` → `hash_b`.

Expected outcome:
- `hash_a` equals `hash_b`.

---

### Interior blank lines preserved

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/b` — version A — with a `# Public` section where `## Interface` has a blank line between content lines:
  ```
  # Public

  ## Interface
  first line

  second line
  ```
- Write `_node.md` for `SPEC/b` — version B — with the same content but no interior blank line:
  ```
  # Public

  ## Interface
  first line
  second line
  ```
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.

Actions:
- Write version A to `SPEC/b/_node.md`. Build `chain` pointing to it. Call `ChainHashCompute(chain)` → `hash_a`.
- Overwrite `SPEC/b/_node.md` with version B. Call `ChainHashCompute(chain)` → `hash_b`.

Expected outcome:
- `hash_a` differs from `hash_b`.

---

## Target

### Target Public and Agent both contribute

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/a` with:
  ```
  # Public

  ## Interface
  some interface

  # Agent
  some agent guidance
  ```
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite `SPEC/a/_node.md`, removing the `# Agent` section entirely (keep only `# Public`).
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### Target without Agent — Agent skipped

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/a` with only:
  ```
  # Public

  ## Interface
  some interface
  ```
  (no `# Agent` section).
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `result`.

Expected outcome:
- No error.
- `len(result)` equals 27.

---

## Input

### Input hashes full file content

Setup:
- Create a temp directory.
- Create an artifact file with content `original input content`.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = ChainItem(unqualified_logical_name="ARTIFACT/input", file_path=<path to artifact file>, qualifier=absent)

Actions:
- Call `ChainHashCompute(chain)` → `hash1`.
- Overwrite the artifact file with `modified input content`.
- Call `ChainHashCompute(chain)` → `hash2`.

Expected outcome:
- `hash1` differs from `hash2`.

---

### No input — skipped

Setup:
- Create a temp directory.
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)` → `result`.

Expected outcome:
- No error.
- `len(result)` equals 27.

---

## Error Cases

### Unreadable spec node file

Setup:
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = []
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to a non-existent file>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)`.

Expected outcome:
- Error `ParseFailure` is raised.

---

### Unreadable artifact file

Setup:
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="ARTIFACT/x", file_path=<path to a non-existent file>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)`.

Expected outcome:
- Error `FileUnreadable` is raised.

---

### Unreadable external file

Setup:
- Write `_node.md` for `SPEC/a` with a `# Public` section containing a `## Interface` subsection with some content.
- Build a `Chain`:
  - `ancestors` = []
  - `dependencies` = [ChainItem(unqualified_logical_name="EXTERNAL/rules.md", file_path=<path to a non-existent file>, qualifier=absent)]
  - `target` = ChainItem(unqualified_logical_name="SPEC/a", file_path=<path to SPEC/a/_node.md>, qualifier=absent)
  - `input` = absent

Actions:
- Call `ChainHashCompute(chain)`.

Expected outcome:
- Error `FileUnreadable` is raised.
