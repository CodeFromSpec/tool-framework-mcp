<!-- code-from-spec: ROOT/functional/logic/chain/resolver@NPMT4z8he7cj9eA260Iq_Xj5HUo -->

## Data structures

```
record ChainItem
  logical_name: string
  file_path: PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  external: list of FrontmatterExternal
  target: ChainItem
  input: optional ChainItem
```

---

## function ChainResolve(target_logical_name: string) -> Chain

**Errors:**
- `UnreadableFrontmatter`: a node's frontmatter cannot be parsed.
- `UnresolvableArtifact`: an `ARTIFACT/` reference's output id does not match any declared output, an `ARTIFACT/` reference is missing an id, or a `depends_on` entry starts with neither `ROOT/` nor `ARTIFACT/`.
- `(LogicalNames.*)`: propagated from `LogicalNameToPath`, `LogicalNameGetParent`.
- `(Frontmatter.*)`: propagated from `FrontmatterParse`.

---

### Step 1 — Resolve ancestors and target

1. If `target_logical_name` equals `"ROOT"`:
   a. Call `LogicalNameToPath("ROOT")` to get the file path.
      If it fails, propagate the error.
   b. Create a `ChainItem` with `logical_name` = `"ROOT"`, `file_path` = the resolved path, `qualifier` = absent.
   c. Set `ancestors` to an empty list.
   d. Set `target` to this `ChainItem`.
   e. Skip to Step 2.

2. Otherwise:
   a. Create an empty list called `name_list`.
   b. Add `target_logical_name` to `name_list`.
   c. Set `current` = `target_logical_name`.
   d. Repeat:
      - Call `LogicalNameGetParent(current)`.
        If it fails, propagate the error.
      - Set `current` to the returned parent name.
      - Add `current` to `name_list`.
      - If `current` equals `"ROOT"`, stop repeating.
   e. Sort `name_list` alphabetically by logical name value.
      This produces root-first order (e.g. `"ROOT"`, `"ROOT/a"`, `"ROOT/a/b"`).
   f. For each name in `name_list`:
      - Call `LogicalNameToPath(name)` to get the file path.
        If it fails, propagate the error.
      - Create a `ChainItem` with that logical name, the resolved file path, and `qualifier` = absent.
   g. The last item in the resulting list is the `target`.
      The remaining items (all but the last) form the `ancestors` list.

---

### Step 2 — Resolve dependencies

1. Call `FrontmatterParse(target.file_path)`.
   If parsing fails, raise error `"unreadable frontmatter"`.
   Store the result as `frontmatter`.

2. Create an empty list called `deps`.

3. For each `ref` in `frontmatter.depends_on`:

   **If `ref` starts with `"ROOT/"`:**
   a. Call `LogicalNameGetQualifier(ref)` to get the qualifier (may be absent).
   b. Call `LogicalNameStripQualifier(ref)` to get the bare logical name.
   c. Call `LogicalNameToPath(bare_logical_name)` to get the file path.
      If it fails, propagate the error.
   d. Create a `ChainItem` with `logical_name` = `bare_logical_name`, `file_path` = resolved path, `qualifier` = the extracted qualifier (absent if none).
   e. Add this `ChainItem` to `deps`.

   **Else if `ref` starts with `"ARTIFACT/"`:**
   a. Call `LogicalNameGetQualifier(ref)` to get the qualifier (the artifact id).
      If the qualifier is absent, raise error `"unresolvable artifact"`.
   b. Call `LogicalNameGetArtifactGenerator(ref)` to get the generating node's logical name.
      If it fails, propagate the error.
   c. Call `LogicalNameToPath(generator_logical_name)` to get the generator node's file path.
      If it fails, propagate the error.
   d. Call `FrontmatterParse(generator_file_path)`.
      If parsing fails, raise error `"unreadable frontmatter"`.
      Store as `generator_frontmatter`.
   e. Search `generator_frontmatter.outputs` for the entry whose `id` equals the qualifier.
      If no match is found, raise error `"unresolvable artifact"`.
      The matching entry's `path` is the artifact file path.
   f. Create a `ChainItem` with `logical_name` = `ref` (the original `ARTIFACT/` reference), `file_path` = the artifact path as `PathCfs`, `qualifier` = the extracted qualifier.
   g. Add this `ChainItem` to `deps`.

   **Else** (starts with neither `"ROOT/"` nor `"ARTIFACT/"`):
   - Raise error `"unresolvable artifact"`.

4. Sort `deps` alphabetically by `file_path.value`, then by `qualifier` (absent sorts before present).

---

### Step 3 — Deduplicate dependencies

1. Create an empty list called `deduped`.

2. For each entry in `deps` (in order):

   **If `LogicalNameIsArtifact(entry.logical_name)` is true (`ARTIFACT/` entry):**
   - Check whether `deduped` already contains an entry with the same `logical_name` (including qualifier) as `entry`.
   - If no such entry exists, add `entry` to `deduped`.
   - If such an entry already exists, discard `entry`.

   **Else (`ROOT/` entry):**
   - If `entry.qualifier` is absent:
     - Check whether `deduped` already contains an entry with the same `file_path` and absent qualifier.
     - If no such entry exists:
       - Add `entry` to `deduped`.
       - Remove any existing entries in `deduped` that have the same `file_path` and a present qualifier (they are now subsumed).
     - If such an entry already exists, discard `entry`.
   - If `entry.qualifier` is present:
     - Check whether `deduped` already contains any entry with the same `file_path` and absent qualifier.
     - If such an unqualified entry exists, discard `entry` (the full section subsumes this subsection).
     - Otherwise, check whether `deduped` already contains an entry with the same `file_path` and the same qualifier value.
       - If yes, discard `entry`.
       - If no, add `entry` to `deduped`.

3. Set `dependencies` = `deduped`.

---

### Step 4 — Collect external

1. Copy `frontmatter.external` into the chain's `external` field as-is, including any `fragments` declarations on each entry.
2. Sort `external` alphabetically by the `path` field of each `FrontmatterExternal` entry.
   Fragments within each entry retain their original declaration order.

---

### Step 5 — Resolve input

1. If `frontmatter.input` is non-empty:
   a. Call `LogicalNameGetQualifier(frontmatter.input)` to get the qualifier (the artifact id).
      If the qualifier is absent, raise error `"unresolvable artifact"`.
   b. Call `LogicalNameGetArtifactGenerator(frontmatter.input)` to get the generating node's logical name.
      If it fails, propagate the error.
   c. Call `LogicalNameToPath(generator_logical_name)` to get the generator node's file path.
      If it fails, propagate the error.
   d. Call `FrontmatterParse(generator_file_path)`.
      If parsing fails, raise error `"unreadable frontmatter"`.
      Store as `input_generator_frontmatter`.
   e. Search `input_generator_frontmatter.outputs` for the entry whose `id` equals the qualifier.
      If no match is found, raise error `"unresolvable artifact"`.
      The matching entry's `path` is the artifact file path.
   f. Create a `ChainItem` with `logical_name` = `frontmatter.input` (the original `ARTIFACT/` reference), `file_path` = the artifact path as `PathCfs`, `qualifier` = the extracted qualifier.
   g. Set `chain.input` = this `ChainItem`.

2. If `frontmatter.input` is empty, set `chain.input` to absent.

---

### Step 6 — Return

Return a `Chain` record with:
- `ancestors` = the ancestors list from Step 1
- `dependencies` = the deduplicated list from Step 3
- `external` = the sorted list from Step 4
- `target` = the target item from Step 1
- `input` = the resolved input item from Step 5, or absent

---

## Contracts

- All file paths in the returned `Chain` are `PathCfs` values (forward slashes, relative to project root).
- File existence is not verified at any step; the caller handles missing files.
- No duplicate entries appear in the `dependencies` list.
- `ancestors` are in root-first alphabetical order.
- `dependencies` are sorted by `file_path` value, then by `qualifier` (absent before present).
- `external` entries are sorted alphabetically by `path`.
- Fragments within each `external` entry retain their original declaration order.
