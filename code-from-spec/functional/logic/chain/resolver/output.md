<!-- code-from-spec: ROOT/functional/logic/chain/resolver@4pPQMKJtbxtDQuTxXrEJgEsJqKc -->

## Namespace

    namespace: chainresolver

## Records

```
record ChainItem
  logical_name: string
  file_path: pathutils.PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  external: list of frontmatter.FrontmatterExternal
  target: ChainItem
  input: optional ChainItem
```

## Functions

---

### ChainResolve

```
function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference's output id does
      not match any declared output, or an ARTIFACT/ reference is
      missing a qualifier, or a depends_on entry has an unrecognized
      prefix.
    - (LogicalNames.*): propagated from LogicalNameToPath,
      LogicalNameGetParent.
    - (Frontmatter.*): propagated from FrontmatterParse.
```

#### Step 1 — Resolve ancestors and target

1. If target_logical_name is "ROOT":
   - Resolve file path using LogicalNameToPath("ROOT").
     If it fails, propagate the error.
   - Create a ChainItem with logical_name "ROOT", the resolved
     file path, and qualifier absent.
   - Set ancestors to an empty list and target to this ChainItem.
   - Proceed to Step 2.

2. Otherwise:
   - Create a list called name_list. Add target_logical_name to it.
   - Repeatedly call LogicalNameGetParent on the last name added,
     appending the result, until the result is "ROOT" (inclusive).
     If LogicalNameGetParent fails, propagate the error.
   - Sort name_list alphabetically by string value.
     This yields root-first order (e.g. "ROOT", "ROOT/a", "ROOT/a/b").
   - For each name in name_list:
     - Call LogicalNameToPath(name). If it fails, propagate the error.
     - Create a ChainItem with that logical_name, the resolved
       file_path, and qualifier absent.
   - The last ChainItem in the sorted list is the target.
   - The remaining ChainItems form the ancestors list (all but the last).

#### Step 2 — Resolve dependencies

1. Call FrontmatterParse(target.file_path).
   If parsing fails, raise error "unreadable frontmatter".
   Store the result as target_frontmatter.

2. Create an empty list called dependencies.

3. For each entry in target_frontmatter.depends_on:

   a. If entry starts with "ROOT/":
      - Call LogicalNameGetQualifier(entry). Store as qualifier
        (may be absent).
      - Call LogicalNameStripQualifier(entry). Store as bare_name.
      - Call LogicalNameToPath(bare_name).
        If it fails, propagate the error.
      - Create a ChainItem with logical_name bare_name, the resolved
        file_path, and qualifier.
      - Append to dependencies.

   b. Else if entry starts with "ARTIFACT/":
      - Call LogicalNameGetQualifier(entry). Store as qualifier.
        If qualifier is absent, raise error "unresolvable artifact".
      - Call LogicalNameGetArtifactGenerator(entry).
        If it fails, propagate the error. Store as generator_name.
      - Call LogicalNameToPath(generator_name).
        If it fails, propagate the error. Store as generator_path.
      - Call FrontmatterParse(generator_path).
        If parsing fails, raise error "unreadable frontmatter".
        Store as generator_frontmatter.
      - Find the output entry in generator_frontmatter.outputs
        whose id equals qualifier.
        If no match is found, raise error "unresolvable artifact".
      - Store the matching output's path as a PathCfs value.
      - Create a ChainItem with logical_name entry (the original
        ARTIFACT/ reference), file_path as the artifact PathCfs,
        and qualifier.
      - Append to dependencies.

   c. Else (entry starts with neither "ROOT/" nor "ARTIFACT/"):
      - Raise error "unresolvable artifact".

4. Sort dependencies alphabetically by file_path value.
   For entries with equal file_path, sort by qualifier:
   absent qualifier sorts before a present qualifier.
   For entries with equal file_path and both qualifiers present,
   sort alphabetically by qualifier value.

#### Step 3 — Deduplicate dependencies

1. Create an empty list called deduped.

2. For each entry in dependencies (in sorted order):

   a. If LogicalNameIsArtifact(entry.logical_name) is true
      (ARTIFACT/ entry):
      - Check whether deduped already contains an entry with the
        same logical_name (including qualifier). Logical names for
        artifacts are always qualified, so equality is exact.
      - If a duplicate exists, skip this entry. Otherwise append it.

   b. Else (ROOT/ entry):
      - If deduped already contains an entry with the same file_path
        and the same qualifier, skip this entry (exact duplicate).
      - If deduped already contains an entry with the same file_path
        and qualifier absent, the full section is already included.
        Skip this entry (subsumed).
      - Otherwise append to deduped.

3. Replace dependencies with deduped.

#### Step 4 — Collect external

1. Copy target_frontmatter.external into the chain's external list.
2. Sort entries alphabetically by path.

#### Step 5 — Resolve input

1. If target_frontmatter.input is empty:
   - Set the chain's input field to absent.

2. Otherwise (input is a non-empty ARTIFACT/ reference):
   - Call LogicalNameGetQualifier(target_frontmatter.input).
     Store as qualifier.
     If qualifier is absent, raise error "unresolvable artifact".
   - Call LogicalNameGetArtifactGenerator(target_frontmatter.input).
     If it fails, propagate the error. Store as generator_name.
   - Call LogicalNameToPath(generator_name).
     If it fails, propagate the error. Store as generator_path.
   - Call FrontmatterParse(generator_path).
     If parsing fails, raise error "unreadable frontmatter".
     Store as generator_frontmatter.
   - Find the output entry in generator_frontmatter.outputs
     whose id equals qualifier.
     If no match is found, raise error "unresolvable artifact".
   - Store the matching output's path as a PathCfs value.
   - Create a ChainItem with logical_name target_frontmatter.input
     (the original ARTIFACT/ reference), file_path as the artifact
     PathCfs, and qualifier.
   - Set the chain's input field to this ChainItem.

#### Return

Return a Chain record with:
- ancestors: the resolved ancestors list (root-first, qualifier absent)
- dependencies: the deduplicated, sorted dependencies list
- external: the sorted external list
- target: the resolved target ChainItem
- input: the resolved input ChainItem, or absent
