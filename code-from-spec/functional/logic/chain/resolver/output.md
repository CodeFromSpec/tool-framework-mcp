<!-- code-from-spec: ROOT/functional/logic/chain/resolver@Yp60IGa5HIWvxxq7L4FcK9XUMbA -->

# Chain Resolver

namespace: chainresolver

## Records

### ChainItem

A single resolved node or artifact in a chain.

```
record ChainItem
  logical_name: string
  file_path:    pathutils.PathCfs
  qualifier:    optional string
```

### Chain

The fully resolved chain for a target node.

```
record Chain
  ancestors:    list of ChainItem
  dependencies: list of ChainItem
  external:     list of frontmatter.FrontmatterExternal
  target:       ChainItem
  input:        optional ChainItem
```

---

## Functions

### ChainResolve

```
function ChainResolve(target_logical_name: string) -> Chain
```

Returns the fully resolved chain for the given target logical name.

**Errors:**
- `UnreadableFrontmatter`: a node's frontmatter cannot be parsed.
- `UnresolvableArtifact`: an `ARTIFACT/` reference has no id, or the
  id does not match any declared output; or a `depends_on` entry
  starts with neither `ROOT/` nor `ARTIFACT/`.
- `(LogicalNames.*)`: propagated from `LogicalNameToPath`,
  `LogicalNameGetParent`.
- `(Frontmatter.*)`: propagated from `FrontmatterParse`.

---

**Step 1 — Resolve ancestors and target**

1. If `target_logical_name` equals `"ROOT"`:
   1. Call `LogicalNameToPath("ROOT")` to get the file path.
      If it fails, propagate the error.
   2. Create a `ChainItem` with:
      - `logical_name` = `"ROOT"`
      - `file_path` = resolved path
      - `qualifier` = absent
   3. Set ancestors to an empty list.
   4. Set target to this item.
   5. Skip to Step 2.

2. Otherwise:
   1. Create an empty list called `name_list`.
   2. Add `target_logical_name` to `name_list`.
   3. Set `current` = `target_logical_name`.
   4. Repeat:
      1. Call `LogicalNameGetParent(current)`.
         If it fails, propagate the error.
      2. Add the returned parent name to `name_list`.
      3. Set `current` = the returned parent name.
      4. If `current` equals `"ROOT"`, stop repeating.
   5. Sort `name_list` alphabetically by logical name value.
      This produces root-first order (e.g. `"ROOT"`, `"ROOT/a"`,
      `"ROOT/a/b"`).
   6. For each name in `name_list`:
      1. Call `LogicalNameToPath(name)` to get the file path.
         If it fails, propagate the error.
      2. Create a `ChainItem` with:
         - `logical_name` = name
         - `file_path` = resolved path
         - `qualifier` = absent
   7. The last item in the resulting list is the **target**.
   8. All preceding items form the **ancestors** list.

---

**Step 2 — Resolve dependencies**

1. Call `FrontmatterParse(target.file_path)`.
   If parsing fails, raise error `"unreadable frontmatter"`.

2. Create an empty list called `deps`.

3. For each entry in `frontmatter.depends_on`:

   **If the entry starts with `"ROOT/"`:**
   1. Call `LogicalNameGetQualifier(entry)` to get the qualifier
      (absent if none).
   2. Call `LogicalNameStripQualifier(entry)` to get the bare
      logical name.
   3. Call `LogicalNameToPath(bare_logical_name)` to get the
      file path. If it fails, propagate the error.
   4. Create a `ChainItem` with:
      - `logical_name` = bare logical name
      - `file_path` = resolved path
      - `qualifier` = qualifier from step 1 (absent if none)
   5. Add the item to `deps`.

   **If the entry starts with `"ARTIFACT/"`:**
   1. Call `LogicalNameGetQualifier(entry)` to get the artifact id.
      If absent, raise error `"unresolvable artifact"`.
   2. Call `LogicalNameGetArtifactGenerator(entry)` to get the
      generating node's logical name. If it fails, propagate the
      error.
   3. Call `LogicalNameToPath(generator_logical_name)` to get the
      generating node's file path. If it fails, propagate the error.
   4. Call `FrontmatterParse(generator_file_path)`.
      If parsing fails, raise error `"unreadable frontmatter"`.
   5. Search the generator's `outputs` list for the entry whose `id`
      equals the artifact id from step 1.
      If no match is found, raise error `"unresolvable artifact"`.
   6. The matching output's `path` is the artifact file path.
      Do not verify existence.
   7. Create a `ChainItem` with:
      - `logical_name` = original entry (the full `ARTIFACT/...` name)
      - `file_path` = artifact file path as `PathCfs`
      - `qualifier` = artifact id from step 1
   8. Add the item to `deps`.

   **Otherwise (neither `ROOT/` nor `ARTIFACT/`):**
   - Raise error `"unresolvable artifact"`.

4. Sort `deps` alphabetically by `file_path` value, then by
   `qualifier` (absent sorts before present).

---

**Step 3 — Deduplicate dependencies**

1. Create an empty list called `deduped`.

2. For each item in `deps` (in order):

   **If the item is an `ARTIFACT/` reference**
   (determined by calling `LogicalNameIsArtifact(item.logical_name)`):
   - Two `ARTIFACT/` entries are duplicates only when they have the
     exact same logical name (including qualifier).
   - If `deduped` already contains an entry with the same
     `logical_name`, skip this item.
   - Otherwise, add it to `deduped`.

   **If the item is a `ROOT/` reference:**
   - If `deduped` already contains an entry with the same `file_path`
     and the same `qualifier`, skip this item (exact duplicate).
   - If `deduped` already contains an entry with the same `file_path`
     and absent qualifier, this item is redundant (the full section
     subsumes any qualifier). Skip this item.
   - Otherwise, add it to `deduped`.
   - After adding, if this item has absent qualifier, remove from
     `deduped` any previously added entries with the same `file_path`
     and a present qualifier (they are now subsumed). Keep this item.

3. Replace `deps` with `deduped`.

---

**Step 4 — Collect external**

1. Copy the `external` list from the target's frontmatter.
2. Sort the entries alphabetically by `path`.
3. Fragments within each entry retain their original declaration order.

---

**Step 5 — Resolve input**

1. If the target's frontmatter `input` field is empty or absent:
   - Set the chain's `input` to absent.
   - Skip to "Return".

2. Otherwise (the `input` field is a non-empty `ARTIFACT/` reference):
   1. Call `LogicalNameGetQualifier(input_value)` to get the
      artifact id. If absent, raise error `"unresolvable artifact"`.
   2. Call `LogicalNameGetArtifactGenerator(input_value)` to get the
      generating node's logical name. If it fails, propagate the
      error.
   3. Call `LogicalNameToPath(generator_logical_name)` to get the
      generating node's file path. If it fails, propagate the error.
   4. Call `FrontmatterParse(generator_file_path)`.
      If parsing fails, raise error `"unreadable frontmatter"`.
   5. Search the generator's `outputs` list for the entry whose `id`
      equals the artifact id from step 1.
      If no match is found, raise error `"unresolvable artifact"`.
   6. The matching output's `path` is the artifact file path.
      Do not verify existence.
   7. Create a `ChainItem` with:
      - `logical_name` = original `input_value` string
      - `file_path` = artifact file path as `PathCfs`
      - `qualifier` = artifact id from step 1
   8. Set the chain's `input` to this item.

---

**Return**

Return a `Chain` record with:
- `ancestors` = ancestors list (root-first order, qualifier absent)
- `dependencies` = deduped deps list (sorted by file path, then qualifier)
- `external` = external list (sorted by path)
- `target` = target item
- `input` = resolved input item, or absent

---

## Contracts

- All file paths in the returned chain are `PathCfs` values (forward
  slashes, relative to project root).
- File existence is not verified at any point.
- The `ancestors` list is in root-first alphabetical order.
- The `dependencies` list has no duplicates and is sorted by
  `file_path` then by `qualifier` (absent before present).
- The `external` list is sorted alphabetically by `path`.
- Fragments within each external entry retain declaration order.
