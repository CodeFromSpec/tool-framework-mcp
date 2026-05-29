<!-- code-from-spec: ROOT/functional/logic/chain/chain_resolver@hfBeWpQWkQdUZS2onkGtJQmYmGk -->

# ChainResolver

## Records

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

## Functions

---

### ChainResolve(target_logical_name: string) -> Chain

  errors:
    - (logical name errors): propagated from LogicalNameToPath, LogicalNameGetParent
    - "unreadable frontmatter": a node's frontmatter cannot be parsed
    - "unresolvable artifact": an ARTIFACT/ reference's output id does not match
      any declared output, or an ARTIFACT/ reference is missing a qualifier

---

#### Step 1 — Resolve ancestors and target

  1. If target_logical_name equals "ROOT":
       Resolve its file path by calling LogicalNameToPath("ROOT").
       If it fails, propagate the error.
       Create a ChainItem with:
         logical_name: "ROOT"
         file_path: resolved path
         qualifier: absent
       Set ancestors to empty list.
       Set target to this ChainItem.
       Skip to Step 2.

  2. Otherwise, build a collection of logical names:
       Add target_logical_name to the collection.
       Set current to target_logical_name.
       Loop:
         Call LogicalNameGetParent(current).
         If it fails, propagate the error.
         Add the parent to the collection.
         Set current to the parent.
         If current equals "ROOT", stop the loop.

  3. Sort the collection alphabetically by logical name.
     This places "ROOT" first and the deepest path last.

  4. For each name in the sorted collection:
       Call LogicalNameToPath(name).
       If it fails, propagate the error.
       Create a ChainItem with:
         logical_name: name
         file_path: resolved path
         qualifier: absent

  5. The last ChainItem in the sorted list is the target.
     All preceding ChainItems form the ancestors list (in sorted order).

---

#### Step 2 — Resolve dependencies

  1. Call FrontmatterParse(target.file_path).
     If parsing fails, raise error "unreadable frontmatter".
     Store the result as target_frontmatter.

  2. For each entry in target_frontmatter.depends_on:

       If the entry starts with "ROOT/":
         a. Call LogicalNameGetQualifier(entry) to extract the qualifier
            (absent if none present).
         b. Call LogicalNameStripQualifier(entry) to get the bare logical name.
         c. Call LogicalNameToPath(bare logical name).
            If it fails, propagate the error.
         d. Create a ChainItem with:
              logical_name: bare logical name
              file_path: resolved path
              qualifier: extracted qualifier (or absent)

       Else if the entry starts with "ARTIFACT/":
         a. Call LogicalNameGetQualifier(entry) to extract the qualifier (artifact id).
            If absent, raise error "unresolvable artifact".
         b. Call LogicalNameGetArtifactGenerator(entry).
            If it fails, propagate the error.
            Store result as generator_logical_name.
         c. Call LogicalNameToPath(generator_logical_name).
            If it fails, propagate the error.
            Store result as generator_file_path.
         d. Call FrontmatterParse(generator_file_path).
            If parsing fails, raise error "unreadable frontmatter".
            Store result as generator_frontmatter.
         e. Search generator_frontmatter.outputs for an entry whose id
            equals the qualifier.
            If no match is found, raise error "unresolvable artifact".
            Store the matching output's path as artifact_path.
         f. Create a ChainItem with:
              logical_name: entry (the original ARTIFACT/ reference)
              file_path: artifact_path (as PathCfs)
              qualifier: the artifact id qualifier

       Else:
         Raise error "unresolvable artifact".

  3. Sort all collected dependency ChainItems alphabetically by file_path value,
     then by qualifier (absent sorts before present).

---

#### Step 3 — Deduplicate dependencies

  1. Initialize an empty result list.

  2. For each ChainItem in the sorted dependencies:

       Call LogicalNameIsArtifact(item.logical_name) to determine its type.

       If it is a ROOT/ entry (not artifact):
         a. If an entry already exists in the result list with the same file_path
            and the same qualifier (including both absent), skip this item.
         b. If an entry already exists in the result list with the same file_path
            and absent qualifier, this item is redundant (the full section already
            includes every subsection). Skip this item.
         c. Otherwise, add this item to the result list.
            If this item has absent qualifier, remove from the result list any
            previously added entries with the same file_path that have a qualifier
            present (they are now subsumed).

       If it is an ARTIFACT/ entry:
         a. If an entry already exists in the result list with the exact same
            logical_name (including qualifier), skip this item.
         b. Otherwise, add this item to the result list.

  3. Replace dependencies with the deduplicated result list.

---

#### Step 4 — Collect external

  1. Copy the external list from target_frontmatter into the chain as-is,
     preserving all FrontmatterExternal records including any fragments declarations.

  2. Sort entries alphabetically by path.
     Fragments within each entry retain their declaration order.

---

#### Step 5 — Resolve input

  1. If target_frontmatter.input is empty or absent:
       Set the chain's input field to absent.

  2. Otherwise (input is a non-empty ARTIFACT/ reference):
       a. Call LogicalNameGetQualifier(input) to extract the qualifier (artifact id).
          If absent, raise error "unresolvable artifact".
       b. Call LogicalNameGetArtifactGenerator(input).
          If it fails, propagate the error.
          Store result as generator_logical_name.
       c. Call LogicalNameToPath(generator_logical_name).
          If it fails, propagate the error.
          Store result as generator_file_path.
       d. Call FrontmatterParse(generator_file_path).
          If parsing fails, raise error "unreadable frontmatter".
          Store result as generator_frontmatter.
       e. Search generator_frontmatter.outputs for an entry whose id
          equals the qualifier.
          If no match is found, raise error "unresolvable artifact".
          Store the matching output's path as artifact_path.
       f. Create a ChainItem with:
            logical_name: input (the original ARTIFACT/ reference)
            file_path: artifact_path (as PathCfs)
            qualifier: the artifact id qualifier
       g. Set the chain's input field to this ChainItem.

---

#### Step 6 — Return

  1. Return a Chain record with:
       ancestors: the ancestors list (root-first order)
       dependencies: the deduplicated, sorted dependencies list
       external: the sorted external list
       target: the target ChainItem
       input: the resolved input ChainItem, or absent

---

## Contracts

- All file paths in the returned Chain are PathCfs values (forward slashes, relative to project root).
- Ancestors are in root-first alphabetical order.
- Dependencies are sorted by file_path value, then by qualifier (absent before present).
- No duplicate entries in the dependencies list.
- External entries are sorted alphabetically by path.
- File existence is not verified; the caller handles missing files.
