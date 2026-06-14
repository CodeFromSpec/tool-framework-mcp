<!-- code-from-spec: ROOT/functional/logic/chain/resolver@k1qrxIE2pNkx_or4KP_qH5vXCTM -->

namespace: chainresolver

record ChainItem
  unqualified_logical_name: string
  file_path: pathutils.PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  target: ChainItem
  input: optional ChainItem


function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference cannot be resolved.
    - (LogicalNames.*): propagated from LogicalNameToPath, LogicalNameGetParent.
    - (Frontmatter.*): propagated from FrontmatterParse.

  Step 1 — Resolve ancestors and target

  1. If target_logical_name is exactly "SPEC":
       Resolve its file path by calling LogicalNameToPath with "SPEC".
       If that fails, propagate the error.
       Create a ChainItem with:
         unqualified_logical_name = "SPEC"
         file_path = resolved path
         qualifier = absent
       Set ancestors = empty list.
       Set target = that ChainItem.
       Go to Step 2.

  2. Otherwise:
       Initialize name_list = empty list.
       Add target_logical_name to name_list.
       Set current = target_logical_name.
       Repeat:
         Call LogicalNameGetParent(current).
         If it fails with NoParent, stop the loop.
         If it fails with any other error, propagate the error.
         Set parent = result.
         Add parent to name_list.
         Set current = parent.
       Sort name_list alphabetically.
       This produces root-first order (e.g. "SPEC", "SPEC/a", "SPEC/a/b").

  3. For each name in name_list:
       Call LogicalNameToPath(name).
       If it fails, propagate the error.
       Create a ChainItem with:
         unqualified_logical_name = name
         file_path = resolved path
         qualifier = absent
       Add it to items_list.

  4. Set target = last item in items_list.
     Set ancestors = all items in items_list except the last.

  Step 2 — Resolve dependencies

  5. Call FrontmatterParse(target.file_path).
     If it fails, raise error "unreadable frontmatter".
     Set target_frontmatter = result.

  6. Initialize dependencies = empty list.

  7. For each entry in target_frontmatter.depends_on:

       If LogicalNameIsSpec(entry) is true:
         Call LogicalNameGetQualifier(entry).
         Set qualifier = result (absent if none present).
         Call LogicalNameStripQualifier(entry).
         Set bare_name = result.
         Call LogicalNameToPath(bare_name).
         If it fails, propagate the error.
         Create a ChainItem with:
           unqualified_logical_name = bare_name
           file_path = resolved path
           qualifier = qualifier
         Add to dependencies.

       Else if LogicalNameIsArtifact(entry) is true:
         Call LogicalNameGetArtifactGenerator(entry).
         If it fails, propagate the error.
         Set generator_name = result.
         Call LogicalNameToPath(generator_name).
         If it fails, propagate the error.
         Set generator_path = result.
         Call FrontmatterParse(generator_path).
         If it fails, raise error "unreadable frontmatter".
         Set generator_frontmatter = result.
         If generator_frontmatter.output is empty:
           Raise error "unresolvable artifact".
         Create a ChainItem with:
           unqualified_logical_name = entry
           file_path = generator_frontmatter.output as PathCfs
           qualifier = absent
         Add to dependencies.

       Else if LogicalNameIsExternal(entry) is true:
         Call LogicalNameExternalToPath(entry).
         If it fails, propagate the error.
         Create a ChainItem with:
           unqualified_logical_name = entry
           file_path = resolved path
           qualifier = absent
         Add to dependencies.

       Else:
         Raise error "unresolvable artifact".

  8. Sort dependencies alphabetically by unqualified_logical_name.
     For entries with the same unqualified_logical_name, sort by qualifier:
       absent qualifier sorts before present qualifier.

  Step 3 — Deduplicate dependencies

  9. Initialize deduped = empty list.

  10. For each item in dependencies (in order):
        Determine its type using LogicalNameIsArtifact, LogicalNameIsSpec,
        and LogicalNameIsExternal.

        If LogicalNameIsSpec(item.unqualified_logical_name) is true:
          Check if deduped already contains an entry with the same
          unqualified_logical_name and qualifier = absent.
          If yes, skip this item (the full section already covers all subsections).
          Check if deduped already contains an entry with the same
          unqualified_logical_name and the same qualifier.
          If yes, skip this item (exact duplicate).
          Otherwise, add item to deduped.
          Additionally, if item.qualifier is absent, remove from deduped any
          previously added entry with the same unqualified_logical_name and
          a non-absent qualifier (they are now redundant).

        If LogicalNameIsArtifact(item.unqualified_logical_name) is true:
          Check if deduped already contains an entry with the same
          unqualified_logical_name.
          If yes, skip this item.
          Otherwise, add item to deduped.

        If LogicalNameIsExternal(item.unqualified_logical_name) is true:
          Check if deduped already contains an entry with the same
          unqualified_logical_name.
          If yes, skip this item.
          Otherwise, add item to deduped.

  11. Set dependencies = deduped.

  Step 4 — Resolve input

  12. If target_frontmatter.input is empty:
        Set chain_input = absent.
        Go to Step 5.

  13. Set input_entry = target_frontmatter.input.

      If LogicalNameIsArtifact(input_entry) is true:
        Call LogicalNameGetArtifactGenerator(input_entry).
        If it fails, propagate the error.
        Set generator_name = result.
        Call LogicalNameToPath(generator_name).
        If it fails, propagate the error.
        Set generator_path = result.
        Call FrontmatterParse(generator_path).
        If it fails, raise error "unreadable frontmatter".
        Set generator_frontmatter = result.
        If generator_frontmatter.output is empty:
          Raise error "unresolvable artifact".
        Create a ChainItem with:
          unqualified_logical_name = input_entry
          file_path = generator_frontmatter.output as PathCfs
          qualifier = absent
        Set chain_input = that ChainItem.

      Else if LogicalNameIsExternal(input_entry) is true:
        Call LogicalNameExternalToPath(input_entry).
        If it fails, propagate the error.
        Create a ChainItem with:
          unqualified_logical_name = input_entry
          file_path = resolved path
          qualifier = absent
        Set chain_input = that ChainItem.

      Else:
        Raise error "unresolvable artifact".

  Step 5 — Return

  14. Return a Chain with:
        ancestors = ancestors
        dependencies = dependencies
        target = target
        input = chain_input
