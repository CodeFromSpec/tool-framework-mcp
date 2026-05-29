<!-- code-from-spec: ROOT/functional/logic/chain/resolver@hfBeWpQWkQdUZS2onkGtJQmYmGk -->

# ChainResolve

## Records

### ChainItem
  - logical_name: string
  - file_path: PathCfs
  - qualifier: optional string

### Chain
  - ancestors: list of ChainItem
  - dependencies: list of ChainItem
  - external: list of FrontmatterExternal
  - target: ChainItem
  - input: optional ChainItem

---

## function ChainResolve(target_logical_name: string) -> Chain

  errors:
    - "unreadable frontmatter": a node's frontmatter cannot be parsed.
    - "unresolvable artifact": an ARTIFACT/ reference's output id does not
      match any declared output, is missing a qualifier, or has an
      unrecognized prefix.
    - (logical name errors): propagated from LogicalNameToPath,
      LogicalNameGetParent, LogicalNameGetArtifactGenerator.

  ### Step 1 — Resolve ancestors and target

  1. If target_logical_name equals "ROOT":
       a. Call LogicalNameToPath("ROOT") to get the file path.
          If it fails, propagate the error.
       b. Create a ChainItem:
            - logical_name: "ROOT"
            - file_path: resolved path
            - qualifier: absent
       c. Set ancestors to empty list.
       d. Set target to this ChainItem.
       e. Proceed to Step 2.

  2. Otherwise:
       a. Initialize name_list as an empty list.
       b. Add target_logical_name to name_list.
       c. Set current_name to target_logical_name.
       d. Repeat:
            i.  Call LogicalNameGetParent(current_name).
                If it fails, propagate the error.
            ii. Add the returned parent name to name_list.
            iii.If the parent name equals "ROOT", stop.
                Otherwise, set current_name to the parent name
                and continue.
       e. Sort name_list alphabetically by logical name value.
          This produces root-first order (e.g. "ROOT", "ROOT/a", "ROOT/a/b").
       f. For each name in name_list:
            i.  Call LogicalNameToPath(name) to get the file path.
                If it fails, propagate the error.
            ii. Create a ChainItem:
                  - logical_name: name
                  - file_path: resolved path
                  - qualifier: absent
       g. The last ChainItem in the sorted list is the target.
          The remaining items (all but the last) form the ancestors list.

  ### Step 2 — Resolve dependencies

  1. Call FrontmatterParse(target.file_path).
     If parsing fails, raise error "unreadable frontmatter".
     Store the result as target_frontmatter.

  2. Initialize dependencies as an empty list.

  3. For each entry in target_frontmatter.depends_on:

       a. If the entry starts with "ROOT/":
            i.   Call LogicalNameGetQualifier(entry).
                 Store result as qualifier (may be absent).
            ii.  Call LogicalNameStripQualifier(entry).
                 Store result as bare_name.
            iii. Call LogicalNameToPath(bare_name).
                 If it fails, propagate the error.
                 Store result as dep_path.
            iv.  Create a ChainItem:
                   - logical_name: bare_name
                   - file_path: dep_path
                   - qualifier: qualifier (from step i)
            v.   Add to dependencies.

       b. Else if the entry starts with "ARTIFACT/":
            i.   Call LogicalNameGetQualifier(entry).
                 If qualifier is absent, raise error "unresolvable artifact".
                 Store qualifier as artifact_id.
            ii.  Call LogicalNameGetArtifactGenerator(entry).
                 If it fails, propagate the error.
                 Store result as generator_name.
            iii. Call LogicalNameToPath(generator_name).
                 If it fails, propagate the error.
                 Store result as generator_path.
            iv.  Call FrontmatterParse(generator_path).
                 If parsing fails, raise error "unreadable frontmatter".
                 Store result as generator_frontmatter.
            v.   Search generator_frontmatter.outputs for an entry
                 whose id equals artifact_id.
                 If none found, raise error "unresolvable artifact".
                 Store the matching output's path as artifact_path.
            vi.  Create a ChainItem:
                   - logical_name: entry (the original ARTIFACT/ reference)
                   - file_path: artifact_path (as PathCfs)
                   - qualifier: artifact_id
            vii. Add to dependencies.

       c. Else:
            Raise error "unresolvable artifact".

  4. Sort dependencies alphabetically by file_path value.
     For entries with the same file_path, sort so that absent qualifier
     comes before present qualifier; if both have qualifiers, sort
     alphabetically by qualifier value.

  ### Step 3 — Deduplicate dependencies

  1. Initialize deduped as an empty list.

  2. For each entry in dependencies (in order):

       a. If LogicalNameIsArtifact(entry.logical_name) is false (ROOT/ entry):
            i.  If deduped already contains an entry with the same
                file_path and the same qualifier, skip this entry.
            ii. If deduped already contains an entry with the same
                file_path and absent qualifier, skip this entry
                (the full section already subsumes any qualified entry).
            iii.Otherwise, add entry to deduped.

       b. If LogicalNameIsArtifact(entry.logical_name) is true (ARTIFACT/ entry):
            i.  If deduped already contains an entry with the same
                logical_name (including qualifier), skip this entry.
            ii. Otherwise, add entry to deduped.

  3. Set dependencies to deduped.

  ### Step 4 — Collect external

  1. Copy target_frontmatter.external into a list.
  2. Sort the list alphabetically by path value.
     Fragments within each entry retain their declaration order.
  3. Store as external.

  ### Step 5 — Resolve input

  1. If target_frontmatter.input is empty:
       Set input to absent.

  2. Otherwise:
       a. Call LogicalNameGetQualifier(target_frontmatter.input).
          If qualifier is absent, raise error "unresolvable artifact".
          Store qualifier as artifact_id.
       b. Call LogicalNameGetArtifactGenerator(target_frontmatter.input).
          If it fails, propagate the error.
          Store result as generator_name.
       c. Call LogicalNameToPath(generator_name).
          If it fails, propagate the error.
          Store result as generator_path.
       d. Call FrontmatterParse(generator_path).
          If parsing fails, raise error "unreadable frontmatter".
          Store result as generator_frontmatter.
       e. Search generator_frontmatter.outputs for an entry
          whose id equals artifact_id.
          If none found, raise error "unresolvable artifact".
          Store the matching output's path as artifact_path.
       f. Create a ChainItem:
            - logical_name: target_frontmatter.input
              (the original ARTIFACT/ reference)
            - file_path: artifact_path (as PathCfs)
            - qualifier: artifact_id
       g. Set input to this ChainItem.

  ### Return

  1. Return a Chain record:
       - ancestors: the ancestors list (root-first order)
       - dependencies: the deduplicated, sorted dependencies list
       - external: the sorted external list
       - target: the target ChainItem
       - input: the resolved input ChainItem, or absent

---

## Contracts

- All file paths in the returned Chain are PathCfs values (forward slashes,
  relative to project root).
- The ancestors list is in root-first alphabetical order.
- The target is not included in the ancestors list.
- The dependencies list contains no duplicates per the deduplication rules.
- The dependencies list is sorted by file_path then qualifier
  (absent before present).
- The external list is sorted alphabetically by path.
- File existence is never verified; resolution is path-based only.
